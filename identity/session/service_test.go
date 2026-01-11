package session

import (
	"context"
	"testing"
	"time"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func (m *MockRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockUserRepository is a mock for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string, tenantID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, username, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string, tenantID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, email, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) ([]*models.User, error) {
	args := m.Called(ctx, tenantID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) (int, error) {
	args := m.Called(ctx, tenantID, filters)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockUserRepository) ListSystem(ctx context.Context, filters *interfaces.UserFilters) ([]*models.User, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) CountSystem(ctx context.Context, filters *interfaces.UserFilters) (int, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockUserRepository) GetByEmailSystem(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetSystemUserByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// TestListSessions tests listing sessions for a user
func TestListSessions(t *testing.T) {
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockUserRepo := new(MockUserRepository)

	userID := uuid.New()
	tenantID := uuid.New()
	otherTenantID := uuid.New()

	user := &models.User{
		ID:       userID,
		Username: "testuser",
	}

	now := time.Now()
	tokens := []*interfaces.RefreshToken{
		{
			ID:        uuid.New(),
			UserID:    userID,
			TenantID:  tenantID,
			CreatedAt: now,
			ExpiresAt: now.Add(24 * time.Hour),
			RevokedAt: nil,
		},
		{
			ID:        uuid.New(),
			UserID:    userID,
			TenantID:  otherTenantID, // Different tenant - should be filtered out
			CreatedAt: now,
			ExpiresAt: now.Add(24 * time.Hour),
			RevokedAt: nil,
		},
		{
			ID:        uuid.New(),
			UserID:    userID,
			TenantID:  tenantID,
			CreatedAt: now,
			ExpiresAt: now.Add(24 * time.Hour),
			RevokedAt: &now, // Revoked - should be filtered out
		},
	}

	mockRefreshTokenRepo.On("GetByUserID", mock.Anything, userID).Return(tokens, nil)
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)

	service := NewService(mockRefreshTokenRepo, mockUserRepo)

	sessions, err := service.ListSessions(context.Background(), userID, tenantID)

	assert.NoError(t, err)
	assert.Len(t, sessions, 1) // Only 1 session should be returned (tenant filtered, revoked filtered)
	assert.Equal(t, tokens[0].ID, sessions[0].ID)
	assert.Equal(t, "testuser", sessions[0].Username)
	mockRefreshTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestListSessions_TenantIsolation tests that sessions are filtered by tenant
func TestListSessions_TenantIsolation(t *testing.T) {
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockUserRepo := new(MockUserRepository)

	userID := uuid.New()
	tenantID := uuid.New()
	otherTenantID := uuid.New()

	user := &models.User{
		ID:       userID,
		Username: "testuser",
	}

	now := time.Now()
	tokens := []*interfaces.RefreshToken{
		{
			ID:        uuid.New(),
			UserID:    userID,
			TenantID:  otherTenantID, // Different tenant
			CreatedAt: now,
			ExpiresAt: now.Add(24 * time.Hour),
			RevokedAt: nil,
		},
	}

	mockRefreshTokenRepo.On("GetByUserID", mock.Anything, userID).Return(tokens, nil)
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)

	service := NewService(mockRefreshTokenRepo, mockUserRepo)

	sessions, err := service.ListSessions(context.Background(), userID, tenantID)

	assert.NoError(t, err)
	assert.Len(t, sessions, 0) // No sessions should be returned (different tenant)
	mockRefreshTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestRevokeSession tests revoking a session
func TestRevokeSession(t *testing.T) {
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockUserRepo := new(MockUserRepository)

	sessionID := uuid.New()

	mockRefreshTokenRepo.On("Revoke", mock.Anything, sessionID).Return(nil)

	service := NewService(mockRefreshTokenRepo, mockUserRepo)

	err := service.RevokeSession(context.Background(), sessionID, "User requested logout")

	assert.NoError(t, err)
	mockRefreshTokenRepo.AssertExpectations(t)
}
