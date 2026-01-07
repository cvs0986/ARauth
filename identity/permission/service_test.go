package permission

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func (m *MockPermissionRepository) GetByName(ctx context.Context, name string, tenantID uuid.UUID) (*models.Permission, error) {
	args := m.Called(ctx, name, tenantID)
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

func (m *MockPermissionRepository) List(ctx context.Context, tenantID uuid.UUID, filters interface{}) ([]*models.Permission, error) {
	args := m.Called(ctx, tenantID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Permission), args.Error(1)
}

func (m *MockPermissionRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func (m *MockPermissionRepository) RemoveFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func (m *MockPermissionRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Permission), args.Error(1)
}

func TestService_Create(t *testing.T) {
	mockRepo := new(MockPermissionRepository)
	service := NewService(mockRepo)

	tenantID := uuid.New()
	req := &CreatePermissionRequest{
		TenantID: tenantID,
		Name:     "users:read",
		Resource: "users",
		Action:   "read",
	}

	expectedPermission := &models.Permission{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     req.Name,
	}

	mockRepo.On("GetByName", mock.Anything, req.Name, tenantID).Return(nil, assert.AnError)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Permission")).Return(nil).Run(func(args mock.Arguments) {
		perm := args.Get(1).(*models.Permission)
		perm.ID = expectedPermission.ID
	})

	permission, err := service.Create(context.Background(), req)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, permission.ID)
	assert.Equal(t, req.Name, permission.Name)

	mockRepo.AssertExpectations(t)
}

func TestService_GetByID(t *testing.T) {
	mockRepo := new(MockPermissionRepository)
	service := NewService(mockRepo)

	permissionID := uuid.New()
	expectedPermission := &models.Permission{
		ID:   permissionID,
		Name: "users:read",
	}

	mockRepo.On("GetByID", mock.Anything, permissionID).Return(expectedPermission, nil)

	permission, err := service.GetByID(context.Background(), permissionID)
	require.NoError(t, err)
	assert.Equal(t, expectedPermission.ID, permission.ID)
	assert.Equal(t, expectedPermission.Name, permission.Name)

	mockRepo.AssertExpectations(t)
}

func TestService_AssignPermissionToRole(t *testing.T) {
	mockRepo := new(MockPermissionRepository)
	service := NewService(mockRepo)

	roleID := uuid.New()
	permissionID := uuid.New()
	expectedPermission := &models.Permission{ID: permissionID}

	mockRepo.On("GetByID", mock.Anything, permissionID).Return(expectedPermission, nil)
	mockRepo.On("AssignPermissionToRole", mock.Anything, roleID, permissionID).Return(nil)

	err := service.AssignPermissionToRole(context.Background(), roleID, permissionID)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

