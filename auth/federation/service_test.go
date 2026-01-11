package federation

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arauth-identity/iam/identity/federation"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockIdentityProviderRepository
type MockIdentityProviderRepository struct {
	mock.Mock
}

func (m *MockIdentityProviderRepository) Create(ctx context.Context, idp *federation.IdentityProvider) error {
	args := m.Called(ctx, idp)
	return args.Error(0)
}

func (m *MockIdentityProviderRepository) GetByID(ctx context.Context, id uuid.UUID) (*federation.IdentityProvider, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*federation.IdentityProvider), args.Error(1)
}

func (m *MockIdentityProviderRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*federation.IdentityProvider, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]*federation.IdentityProvider), args.Error(1)
}

func (m *MockIdentityProviderRepository) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*federation.IdentityProvider, error) {
	args := m.Called(ctx, tenantID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*federation.IdentityProvider), args.Error(1)
}

func (m *MockIdentityProviderRepository) Update(ctx context.Context, idp *federation.IdentityProvider) error {
	args := m.Called(ctx, idp)
	return args.Error(0)
}

func (m *MockIdentityProviderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestVerifyIdentityProvider_OIDC(t *testing.T) {
	// Setup mock OIDC server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"issuer":                 "http://example.com",
				"authorization_endpoint": "http://example.com/auth",
				"token_endpoint":         "http://example.com/token",
				"userinfo_endpoint":      "http://example.com/userinfo",
				"jwks_uri":               "http://example.com/jwks",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	mockRepo := &MockIdentityProviderRepository{}
	// Only dependency needed for VerifyIdentityProvider is idpRepo
	service := NewService(mockRepo, nil, nil, nil, nil, nil)

	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		idp := &federation.IdentityProvider{
			ID:   id,
			Type: federation.IdentityProviderTypeOIDC,
			Configuration: map[string]interface{}{
				"issuer_url":    server.URL,
				"client_id":     "client-id",
				"client_secret": "client-secret",
			},
		}

		mockRepo.On("GetByID", mock.Anything, id).Return(idp, nil).Once()

		result, err := service.VerifyIdentityProvider(context.Background(), id)
		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "OIDC provider verified successfully", result.Message)
	})

	t.Run("discovery_failure", func(t *testing.T) {
		id := uuid.New()
		idp := &federation.IdentityProvider{
			ID:   id,
			Type: federation.IdentityProviderTypeOIDC,
			Configuration: map[string]interface{}{
				"issuer_url":    "http://localhost:54321", // Invalid port
				"client_id":     "client-id",
				"client_secret": "client-secret",
			},
		}

		mockRepo.On("GetByID", mock.Anything, id).Return(idp, nil).Once()

		result, err := service.VerifyIdentityProvider(context.Background(), id)
		assert.NoError(t, err) // Verification failure is not an error, it's a result
		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "Failed to discover OIDC provider")
	})
}

func TestVerifyIdentityProvider_SAML(t *testing.T) {
	mockRepo := &MockIdentityProviderRepository{}
	service := NewService(mockRepo, nil, nil, nil, nil, nil)

	// Generate a valid certificate
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test Cert",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}
	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	certBlock := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certPEM := string(certBlock)

	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		idp := &federation.IdentityProvider{
			ID:   id,
			Type: federation.IdentityProviderTypeSAML,
			Configuration: map[string]interface{}{
				"entity_id":        "urn:example:idp",
				"sso_url":          "http://example.com/sso",
				"x509_certificate": certPEM,
			},
		}

		mockRepo.On("GetByID", mock.Anything, id).Return(idp, nil).Once()

		result, err := service.VerifyIdentityProvider(context.Background(), id)
		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "SAML provider configuration verified", result.Message)
	})

	t.Run("invalid_certificate", func(t *testing.T) {
		id := uuid.New()
		idp := &federation.IdentityProvider{
			ID:   id,
			Type: federation.IdentityProviderTypeSAML,
			Configuration: map[string]interface{}{
				"entity_id":        "urn:example:idp",
				"sso_url":          "http://example.com/sso",
				"x509_certificate": "INVALID_PEM",
			},
		}

		mockRepo.On("GetByID", mock.Anything, id).Return(idp, nil).Once()

		result, err := service.VerifyIdentityProvider(context.Background(), id)
		assert.NoError(t, err)
		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "Invalid X509 Certificate")
	})

	t.Run("expired_certificate", func(t *testing.T) {
		// Generate expired cert
		expiredTemplate := template
		expiredTemplate.NotBefore = time.Now().Add(-2 * time.Hour)
		expiredTemplate.NotAfter = time.Now().Add(-1 * time.Hour)
		derBytes, _ := x509.CreateCertificate(rand.Reader, &expiredTemplate, &expiredTemplate, &priv.PublicKey, priv)
		expiredCertPEM := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}))

		id := uuid.New()
		idp := &federation.IdentityProvider{
			ID:   id,
			Type: federation.IdentityProviderTypeSAML,
			Configuration: map[string]interface{}{
				"entity_id":        "urn:example:idp",
				"sso_url":          "http://example.com/sso",
				"x509_certificate": expiredCertPEM,
			},
		}

		mockRepo.On("GetByID", mock.Anything, id).Return(idp, nil).Once()

		result, err := service.VerifyIdentityProvider(context.Background(), id)
		assert.NoError(t, err)
		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "Certificate has expired")
	})
}
