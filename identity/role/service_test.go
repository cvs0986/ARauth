package role

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/storage/interfaces"
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

func (m *MockRoleRepository) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Role, error) {
	args := m.Called(ctx, tenantID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockRoleRepository) Update(ctx context.Context, r *models.Role) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *MockRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoleRepository) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.RoleFilters) ([]*models.Role, error) {
	args := m.Called(ctx, tenantID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Role), args.Error(1)
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
	mockPermRepo := new(MockPermissionRepository)
	service := NewService(mockRepo, mockPermRepo)

	tenantID := uuid.New()
	desc := "Administrator role"
	req := &CreateRoleRequest{
		TenantID:    tenantID,
		Name:        "admin",
		Description: &desc,
	}

	expectedRole := &models.Role{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     req.Name,
	}

	mockRepo.On("GetByName", mock.Anything, tenantID, req.Name).Return(nil, assert.AnError)
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
	mockPermRepo := new(MockPermissionRepository)
	service := NewService(mockRepo, mockPermRepo)

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

func TestService_AssignRoleToUser(t *testing.T) {
	mockRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	service := NewService(mockRepo, mockPermRepo)

	userID := uuid.New()
	roleID := uuid.New()
	expectedRole := &models.Role{ID: roleID}

	mockRepo.On("GetByID", mock.Anything, roleID).Return(expectedRole, nil)
	mockRepo.On("AssignRoleToUser", mock.Anything, userID, roleID).Return(nil)

	err := service.AssignRoleToUser(context.Background(), userID, roleID)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

// MockPermissionRepository for role service tests
type MockPermissionRepository struct {
	mock.Mock
}

func (m *MockPermissionRepository) Create(ctx context.Context, p *models.Permission) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPermissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Permission), args.Error(1)
}

func (m *MockPermissionRepository) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Permission, error) {
	args := m.Called(ctx, tenantID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Permission), args.Error(1)
}

func (m *MockPermissionRepository) Update(ctx context.Context, p *models.Permission) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPermissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPermissionRepository) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.PermissionFilters) ([]*models.Permission, error) {
	args := m.Called(ctx, tenantID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Permission), args.Error(1)
}

func (m *MockPermissionRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Permission), args.Error(1)
}

func (m *MockPermissionRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func (m *MockPermissionRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

