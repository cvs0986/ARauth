package testutil

import (
	"context"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, u *models.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
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

func (m *MockUserRepository) Update(ctx context.Context, u *models.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
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
	return args.Int(0), args.Error(1)
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

