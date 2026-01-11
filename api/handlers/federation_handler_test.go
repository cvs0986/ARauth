package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arauth-identity/iam/auth/federation"
	idf "github.com/arauth-identity/iam/identity/federation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFederationService
type MockFederationService struct {
	mock.Mock
}

func (m *MockFederationService) CreateIdentityProvider(ctx context.Context, tenantID uuid.UUID, req *federation.CreateIdPRequest) (*idf.IdentityProvider, error) {
	args := m.Called(ctx, tenantID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*idf.IdentityProvider), args.Error(1)
}

func (m *MockFederationService) GetIdentityProvider(ctx context.Context, id uuid.UUID) (*idf.IdentityProvider, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*idf.IdentityProvider), args.Error(1)
}

func (m *MockFederationService) GetIdentityProvidersByTenant(ctx context.Context, tenantID uuid.UUID) ([]*idf.IdentityProvider, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]*idf.IdentityProvider), args.Error(1)
}

func (m *MockFederationService) UpdateIdentityProvider(ctx context.Context, id uuid.UUID, req *federation.UpdateIdPRequest) (*idf.IdentityProvider, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*idf.IdentityProvider), args.Error(1)
}

func (m *MockFederationService) DeleteIdentityProvider(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFederationService) VerifyIdentityProvider(ctx context.Context, id uuid.UUID) (*federation.VerificationResult, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*federation.VerificationResult), args.Error(1)
}

func (m *MockFederationService) InitiateOIDCLogin(ctx context.Context, tenantID uuid.UUID, providerID uuid.UUID, redirectURI string) (string, string, error) {
	args := m.Called(ctx, tenantID, providerID, redirectURI)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockFederationService) HandleOIDCCallback(ctx context.Context, providerID uuid.UUID, code, state, redirectURI string) (*federation.LoginResponse, error) {
	args := m.Called(ctx, providerID, code, state, redirectURI)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*federation.LoginResponse), args.Error(1)
}

func (m *MockFederationService) InitiateSAMLLogin(ctx context.Context, tenantID uuid.UUID, providerID uuid.UUID, acsURL string) (string, error) {
	args := m.Called(ctx, tenantID, providerID, acsURL)
	return args.String(0), args.Error(1)
}

func (m *MockFederationService) HandleSAMLCallback(ctx context.Context, providerID uuid.UUID, samlResponse, relayState string) (*federation.LoginResponse, error) {
	args := m.Called(ctx, providerID, samlResponse, relayState)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*federation.LoginResponse), args.Error(1)
}

func TestFederationHandler_VerifyIdentityProvider(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tenantID := uuid.New()
	providerID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockService := &MockFederationService{}
		handler := NewFederationHandler(mockService)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("tenant_id", tenantID)
			c.Next()
		})
		router.POST("/api/v1/identity-providers/:id/verify", handler.VerifyIdentityProvider)

		provider := &idf.IdentityProvider{ID: providerID, TenantID: tenantID}
		mockService.On("GetIdentityProvider", mock.Anything, providerID).Return(provider, nil)
		mockService.On("VerifyIdentityProvider", mock.Anything, providerID).Return(&federation.VerificationResult{
			Success: true,
			Message: "Verified",
		}, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/identity-providers/"+providerID.String()+"/verify", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Verified")
	})

	t.Run("verification_failure", func(t *testing.T) {
		mockService := &MockFederationService{}
		handler := NewFederationHandler(mockService)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("tenant_id", tenantID)
			c.Next()
		})
		router.POST("/api/v1/identity-providers/:id/verify", handler.VerifyIdentityProvider)

		provider := &idf.IdentityProvider{ID: providerID, TenantID: tenantID}
		mockService.On("GetIdentityProvider", mock.Anything, providerID).Return(provider, nil)
		mockService.On("VerifyIdentityProvider", mock.Anything, providerID).Return(&federation.VerificationResult{
			Success: false,
			Message: "Failed",
		}, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/identity-providers/"+providerID.String()+"/verify", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Failed")
	})

	t.Run("forbidden", func(t *testing.T) {
		mockService := &MockFederationService{}
		handler := NewFederationHandler(mockService)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("tenant_id", tenantID)
			c.Next()
		})
		router.POST("/api/v1/identity-providers/:id/verify", handler.VerifyIdentityProvider)

		otherTenantID := uuid.New()
		provider := &idf.IdentityProvider{ID: providerID, TenantID: otherTenantID}
		mockService.On("GetIdentityProvider", mock.Anything, providerID).Return(provider, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/identity-providers/"+providerID.String()+"/verify", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("not_found", func(t *testing.T) {
		mockService := &MockFederationService{}
		handler := NewFederationHandler(mockService)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("tenant_id", tenantID)
			c.Next()
		})
		router.POST("/api/v1/identity-providers/:id/verify", handler.VerifyIdentityProvider)

		mockService.On("GetIdentityProvider", mock.Anything, providerID).Return(nil, errors.New("not found"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/identity-providers/"+providerID.String()+"/verify", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
