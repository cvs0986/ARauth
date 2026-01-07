package role

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/storage/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func (m *MockRoleRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.Role, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Role), args.Error(1)
}

func (m *MockRoleRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockRoleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

// MockPermissionRepository is a mock implementation of PermissionRepository
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

func TestService_Create_EmptyName(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	service := NewService(mockRoleRepo, mockPermRepo)

	req := &CreateRoleRequest{
		TenantID: uuid.New(),
		Name:     "", // Empty name
	}

	_, err := service.Create(context.Background(), req)
	assert.Error(t, err)
}

func TestService_Create_DuplicateName(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	service := NewService(mockRoleRepo, mockPermRepo)

	tenantID := uuid.New()
	roleName := "Admin"

	// Mock existing role
	mockRoleRepo.On("GetByName", mock.Anything, tenantID, roleName).Return(&struct {
		ID       uuid.UUID
		TenantID uuid.UUID
		Name     string
	}{ID: uuid.New(), TenantID: tenantID, Name: roleName}, nil)

	req := &CreateRoleRequest{
		TenantID: tenantID,
		Name:     roleName,
	}

	_, err := service.Create(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	mockRoleRepo.AssertExpectations(t)
}

func TestService_GetByID_NotFound(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	service := NewService(mockRoleRepo, mockPermRepo)

	nonExistentID := uuid.New()
	mockRoleRepo.On("GetByID", mock.Anything, nonExistentID).Return(nil, assert.AnError)

	_, err := service.GetByID(context.Background(), nonExistentID)
	assert.Error(t, err)
	mockRoleRepo.AssertExpectations(t)
}

func TestService_AssignRoleToUser_RoleNotFound(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	service := NewService(mockRoleRepo, mockPermRepo)

	userID := uuid.New()
	roleID := uuid.New()

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(nil, assert.AnError)

	err := service.AssignRoleToUser(context.Background(), userID, roleID)
	assert.Error(t, err)
	mockRoleRepo.AssertExpectations(t)
}

func TestService_AssignPermissionToRole_PermissionNotFound(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	service := NewService(mockRoleRepo, mockPermRepo)

	roleID := uuid.New()
	permissionID := uuid.New()

	mockPermRepo.On("GetByID", mock.Anything, permissionID).Return(nil, assert.AnError)

	err := service.AssignPermissionToRole(context.Background(), roleID, permissionID)
	assert.Error(t, err)
	mockPermRepo.AssertExpectations(t)
}

