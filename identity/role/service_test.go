package role

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRoleRepository is a mock implementation of RoleRepository
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) Create(ctx context.Context, r *models.Role) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *MockRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleRepository) GetByName(ctx context.Context, name string, tenantID uuid.UUID) (*models.Role, error) {
	args := m.Called(ctx, name, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleRepository) Update(ctx context.Context, r *models.Role) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *MockRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoleRepository) List(ctx context.Context, tenantID uuid.UUID, filters interface{}) ([]*models.Role, error) {
	args := m.Called(ctx, tenantID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Role), args.Error(1)
}

func (m *MockRoleRepository) AssignToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockRoleRepository) RemoveFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockRoleRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.Role, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Role), args.Error(1)
}

func TestService_Create(t *testing.T) {
	mockRepo := new(MockRoleRepository)
	service := NewService(mockRepo)

	tenantID := uuid.New()
	req := &CreateRoleRequest{
		TenantID:    tenantID,
		Name:        "admin",
		Description: "Administrator role",
	}

	expectedRole := &models.Role{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     req.Name,
	}

	mockRepo.On("GetByName", mock.Anything, req.Name, tenantID).Return(nil, assert.AnError)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Role")).Return(nil).Run(func(args mock.Arguments) {
		role := args.Get(1).(*models.Role)
		role.ID = expectedRole.ID
	})

	role, err := service.Create(context.Background(), req)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, role.ID)
	assert.Equal(t, req.Name, role.Name)

	mockRepo.AssertExpectations(t)
}

func TestService_GetByID(t *testing.T) {
	mockRepo := new(MockRoleRepository)
	service := NewService(mockRepo)

	roleID := uuid.New()
	expectedRole := &models.Role{
		ID:   roleID,
		Name: "admin",
	}

	mockRepo.On("GetByID", mock.Anything, roleID).Return(expectedRole, nil)

	role, err := service.GetByID(context.Background(), roleID)
	require.NoError(t, err)
	assert.Equal(t, expectedRole.ID, role.ID)
	assert.Equal(t, expectedRole.Name, role.Name)

	mockRepo.AssertExpectations(t)
}

func TestService_AssignToUser(t *testing.T) {
	mockRepo := new(MockRoleRepository)
	service := NewService(mockRepo)

	userID := uuid.New()
	roleID := uuid.New()

	mockRepo.On("AssignToUser", mock.Anything, userID, roleID).Return(nil)

	err := service.AssignToUser(context.Background(), userID, roleID)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

