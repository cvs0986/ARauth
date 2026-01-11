package oauthclient

import (
	"context"
	"testing"

	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockOAuthClientRepository is a mock for testing
type MockOAuthClientRepository struct {
	mock.Mock
}

func (m *MockOAuthClientRepository) Create(ctx context.Context, client *interfaces.OAuthClient) error {
	args := m.Called(ctx, client)
	return args.Error(0)
}

func (m *MockOAuthClientRepository) GetByID(ctx context.Context, id uuid.UUID) (*interfaces.OAuthClient, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.OAuthClient), args.Error(1)
}

func (m *MockOAuthClientRepository) GetByClientID(ctx context.Context, clientID string) (*interfaces.OAuthClient, error) {
	args := m.Called(ctx, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.OAuthClient), args.Error(1)
}

func (m *MockOAuthClientRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*interfaces.OAuthClient, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*interfaces.OAuthClient), args.Error(1)
}

func (m *MockOAuthClientRepository) Update(ctx context.Context, client *interfaces.OAuthClient) error {
	args := m.Called(ctx, client)
	return args.Error(0)
}

func (m *MockOAuthClientRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockRefreshTokenRepository is a mock for testing
type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *interfaces.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*interfaces.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*interfaces.RefreshToken, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*interfaces.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) Revoke(ctx context.Context, tokenID uuid.UUID) error {
	args := m.Called(ctx, tokenID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	args := m.Called(ctx, tokenHash)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) RevokeByClientID(ctx context.Context, clientID string) (int, error) {
	args := m.Called(ctx, clientID)
	return args.Int(0), args.Error(1)
}

func (m *MockRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// TestCreateClient_Success tests successful client creation
func TestCreateClient_Success(t *testing.T) {
	mockRepo := new(MockOAuthClientRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	service := NewService(mockRepo, mockRefreshTokenRepo)

	tenantID := uuid.New()
	createdBy := uuid.New()

	req := &CreateClientRequest{
		Name:           "Test Client",
		Description:    "Test Description",
		RedirectURIs:   []string{"https://example.com/callback"},
		GrantTypes:     []string{"authorization_code"},
		Scopes:         []string{"openid", "profile"},
		IsConfidential: true,
	}

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(client *interfaces.OAuthClient) bool {
		return client.TenantID == tenantID && client.Name == "Test Client"
	})).Return(nil)

	resp, err := service.CreateClient(context.Background(), tenantID, req, createdBy)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.ClientID)
	assert.NotEmpty(t, resp.ClientSecret) // One-time secret returned
	assert.Equal(t, "Test Client", resp.Name)
	assert.True(t, resp.IsConfidential)
	mockRepo.AssertExpectations(t)
}

// TestCreateClient_SecretHashed tests that secrets are bcrypt hashed
func TestCreateClient_SecretHashed(t *testing.T) {
	mockRepo := new(MockOAuthClientRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	service := NewService(mockRepo, mockRefreshTokenRepo)

	tenantID := uuid.New()
	createdBy := uuid.New()

	req := &CreateClientRequest{
		Name:           "Test Client",
		Description:    "Test",
		RedirectURIs:   []string{"https://example.com/callback"},
		GrantTypes:     []string{"authorization_code"},
		Scopes:         []string{"openid"},
		IsConfidential: true,
	}

	var capturedHash string
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(client *interfaces.OAuthClient) bool {
		capturedHash = client.ClientSecretHash
		return true
	})).Return(nil)

	resp, err := service.CreateClient(context.Background(), tenantID, req, createdBy)

	assert.NoError(t, err)
	assert.NotEmpty(t, capturedHash)

	// Verify the hash is valid bcrypt and matches the returned secret
	err = bcrypt.CompareHashAndPassword([]byte(capturedHash), []byte(resp.ClientSecret))
	assert.NoError(t, err, "Secret hash should match the returned secret")

	mockRepo.AssertExpectations(t)
}

// TestListClients_NoSecrets tests that secrets are not included in list
func TestListClients_NoSecrets(t *testing.T) {
	mockRepo := new(MockOAuthClientRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	service := NewService(mockRepo, mockRefreshTokenRepo)

	tenantID := uuid.New()
	desc := "Test Description"

	repoClients := []*interfaces.OAuthClient{
		{
			ID:               uuid.New(),
			TenantID:         tenantID,
			Name:             "Client 1",
			ClientID:         "client_abc123",
			ClientSecretHash: "bcrypt_hash_here",
			Description:      &desc,
			RedirectURIs:     []string{"https://example.com"},
			GrantTypes:       []string{"authorization_code"},
			Scopes:           []string{"openid"},
			IsConfidential:   true,
			IsActive:         true,
		},
	}

	mockRepo.On("ListByTenant", mock.Anything, tenantID).Return(repoClients, nil)

	clients, err := service.ListClients(context.Background(), tenantID)

	assert.NoError(t, err)
	assert.Len(t, clients, 1)
	assert.Equal(t, "Client 1", clients[0].Name)
	assert.Equal(t, "client_abc123", clients[0].ClientID)
	// Client model doesn't have ClientSecretHash field - secret is never exposed
	mockRepo.AssertExpectations(t)
}

// TestGetClient_NoSecret tests that secret is not included in get
func TestGetClient_NoSecret(t *testing.T) {
	mockRepo := new(MockOAuthClientRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	service := NewService(mockRepo, mockRefreshTokenRepo)

	clientID := uuid.New()
	tenantID := uuid.New()
	desc := "Test"

	repoClient := &interfaces.OAuthClient{
		ID:               clientID,
		TenantID:         tenantID,
		Name:             "Test Client",
		ClientID:         "client_xyz789",
		ClientSecretHash: "bcrypt_hash_here",
		Description:      &desc,
		RedirectURIs:     []string{"https://example.com"},
		GrantTypes:       []string{"authorization_code"},
		Scopes:           []string{"openid"},
		IsConfidential:   true,
		IsActive:         true,
	}

	mockRepo.On("GetByID", mock.Anything, clientID).Return(repoClient, nil)

	client, err := service.GetClient(context.Background(), clientID, tenantID)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "Test Client", client.Name)
	// Client model doesn't have ClientSecretHash field - secret is never exposed
	mockRepo.AssertExpectations(t)
}

// TestGetClient_TenantIsolation tests cross-tenant access is denied
func TestGetClient_TenantIsolation(t *testing.T) {
	mockRepo := new(MockOAuthClientRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	service := NewService(mockRepo, mockRefreshTokenRepo)

	clientID := uuid.New()
	clientTenantID := uuid.New()
	requestTenantID := uuid.New() // Different tenant

	repoClient := &interfaces.OAuthClient{
		ID:       clientID,
		TenantID: clientTenantID, // Client belongs to different tenant
		Name:     "Test Client",
	}

	mockRepo.On("GetByID", mock.Anything, clientID).Return(repoClient, nil)

	client, err := service.GetClient(context.Background(), clientID, requestTenantID)

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "does not belong to tenant")
	mockRepo.AssertExpectations(t)
}

// TestRotateSecret_Success tests successful secret rotation
func TestRotateSecret_Success(t *testing.T) {
	mockRepo := new(MockOAuthClientRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	service := NewService(mockRepo, mockRefreshTokenRepo)

	clientID := uuid.New()
	tenantID := uuid.New()

	repoClient := &interfaces.OAuthClient{
		ID:               clientID,
		TenantID:         tenantID,
		ClientID:         "client_abc123",
		ClientSecretHash: "old_bcrypt_hash",
	}

	mockRepo.On("GetByID", mock.Anything, clientID).Return(repoClient, nil)

	var newHash string
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(client *interfaces.OAuthClient) bool {
		newHash = client.ClientSecretHash
		return client.ID == clientID && client.ClientSecretHash != "old_bcrypt_hash"
	})).Return(nil)

	// Mock token revocation (Phase B4.1)
	mockRefreshTokenRepo.On("RevokeByClientID", mock.Anything, "client_abc123").Return(2, nil)

	resp, err := service.RotateSecret(context.Background(), clientID, tenantID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.ClientSecret) // New one-time secret returned
	assert.Equal(t, "client_abc123", resp.ClientID)
	assert.NotNil(t, resp.RevokedTokens)
	assert.Equal(t, 2, *resp.RevokedTokens)

	// Verify new hash is different from old hash
	assert.NotEqual(t, "old_bcrypt_hash", newHash)

	// Verify new hash matches new secret
	err = bcrypt.CompareHashAndPassword([]byte(newHash), []byte(resp.ClientSecret))
	assert.NoError(t, err, "New secret hash should match the returned secret")

	mockRepo.AssertExpectations(t)
	mockRefreshTokenRepo.AssertExpectations(t)
}

// TestRotateSecret_TenantIsolation tests cross-tenant rotation is denied
func TestRotateSecret_TenantIsolation(t *testing.T) {
	mockRepo := new(MockOAuthClientRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	service := NewService(mockRepo, mockRefreshTokenRepo)

	clientID := uuid.New()
	clientTenantID := uuid.New()
	requestTenantID := uuid.New() // Different tenant

	repoClient := &interfaces.OAuthClient{
		ID:       clientID,
		TenantID: clientTenantID, // Client belongs to different tenant
	}

	mockRepo.On("GetByID", mock.Anything, clientID).Return(repoClient, nil)

	resp, err := service.RotateSecret(context.Background(), clientID, requestTenantID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "does not belong to tenant")
	mockRepo.AssertExpectations(t)
}

// TestDeleteClient_Success tests successful client deletion
func TestDeleteClient_Success(t *testing.T) {
	mockRepo := new(MockOAuthClientRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	service := NewService(mockRepo, mockRefreshTokenRepo)

	clientID := uuid.New()
	tenantID := uuid.New()

	repoClient := &interfaces.OAuthClient{
		ID:       clientID,
		TenantID: tenantID,
	}

	mockRepo.On("GetByID", mock.Anything, clientID).Return(repoClient, nil)
	mockRepo.On("Delete", mock.Anything, clientID).Return(nil)

	err := service.DeleteClient(context.Background(), clientID, tenantID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestDeleteClient_TenantIsolation tests cross-tenant deletion is denied
func TestDeleteClient_TenantIsolation(t *testing.T) {
	mockRepo := new(MockOAuthClientRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	service := NewService(mockRepo, mockRefreshTokenRepo)

	clientID := uuid.New()
	clientTenantID := uuid.New()
	requestTenantID := uuid.New() // Different tenant

	repoClient := &interfaces.OAuthClient{
		ID:       clientID,
		TenantID: clientTenantID, // Client belongs to different tenant
	}

	mockRepo.On("GetByID", mock.Anything, clientID).Return(repoClient, nil)

	err := service.DeleteClient(context.Background(), clientID, requestTenantID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not belong to tenant")
	mockRepo.AssertExpectations(t)
}
