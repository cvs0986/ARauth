package permission

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/storage/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
	mockRepo := new(MockPermissionRepository)
	service := NewService(mockRepo)

	req := &CreatePermissionRequest{
		TenantID: uuid.New(),
		Name:     "", // Empty name
		Resource: "users",
		Action:   "read",
	}

	_, err := service.Create(context.Background(), req)
	assert.Error(t, err)
}

func TestService_Create_EmptyResource(t *testing.T) {
	mockRepo := new(MockPermissionRepository)
	service := NewService(mockRepo)

	req := &CreatePermissionRequest{
		TenantID: uuid.New(),
		Name:     "users:read",
		Resource: "", // Empty resource
		Action:   "read",
	}

	_, err := service.Create(context.Background(), req)
	assert.Error(t, err)
}

func TestService_Create_EmptyAction(t *testing.T) {
	mockRepo := new(MockPermissionRepository)
	service := NewService(mockRepo)

	req := &CreatePermissionRequest{
		TenantID: uuid.New(),
		Name:     "users:read",
		Resource: "users",
		Action:   "", // Empty action
	}

	_, err := service.Create(context.Background(), req)
	assert.Error(t, err)
}

func TestService_Create_DuplicateName(t *testing.T) {
	mockRepo := new(MockPermissionRepository)
	service := NewService(mockRepo)

	tenantID := uuid.New()
	permissionName := "users:read"

	// Mock existing permission
	mockRepo.On("GetByName", mock.Anything, tenantID, permissionName).Return(&struct {
		ID       uuid.UUID
		TenantID uuid.UUID
		Name     string
	}{ID: uuid.New(), TenantID: tenantID, Name: permissionName}, nil)

	req := &CreatePermissionRequest{
		TenantID: tenantID,
		Name:     permissionName,
		Resource: "users",
		Action:   "read",
	}

	_, err := service.Create(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	mockRepo.AssertExpectations(t)
}

func TestService_GetByID_NotFound(t *testing.T) {
	mockRepo := new(MockPermissionRepository)
	service := NewService(mockRepo)

	nonExistentID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, nonExistentID).Return(nil, assert.AnError)

	_, err := service.GetByID(context.Background(), nonExistentID)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

