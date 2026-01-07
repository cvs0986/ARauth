package role

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Create_EmptyName(t *testing.T) {
	mockRoleRepo := testutil.NewMockRoleRepository()
	mockPermRepo := testutil.NewMockPermissionRepository()
	service := NewService(mockRoleRepo, mockPermRepo)

	req := &CreateRoleRequest{
		TenantID: uuid.New(),
		Name:     "", // Empty name
	}

	_, err := service.Create(context.Background(), req)
	assert.Error(t, err)
}

func TestService_Create_DuplicateName(t *testing.T) {
	mockRoleRepo := testutil.NewMockRoleRepository()
	mockPermRepo := testutil.NewMockPermissionRepository()
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
	mockRoleRepo := testutil.NewMockRoleRepository()
	mockPermRepo := testutil.NewMockPermissionRepository()
	service := NewService(mockRoleRepo, mockPermRepo)

	nonExistentID := uuid.New()
	mockRoleRepo.On("GetByID", mock.Anything, nonExistentID).Return(nil, assert.AnError)

	_, err := service.GetByID(context.Background(), nonExistentID)
	assert.Error(t, err)
	mockRoleRepo.AssertExpectations(t)
}

func TestService_AssignRoleToUser_RoleNotFound(t *testing.T) {
	mockRoleRepo := testutil.NewMockRoleRepository()
	mockPermRepo := testutil.NewMockPermissionRepository()
	service := NewService(mockRoleRepo, mockPermRepo)

	userID := uuid.New()
	roleID := uuid.New()

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(nil, assert.AnError)

	err := service.AssignRoleToUser(context.Background(), userID, roleID)
	assert.Error(t, err)
	mockRoleRepo.AssertExpectations(t)
}

func TestService_AssignPermissionToRole_PermissionNotFound(t *testing.T) {
	mockRoleRepo := testutil.NewMockRoleRepository()
	mockPermRepo := testutil.NewMockPermissionRepository()
	service := NewService(mockRoleRepo, mockPermRepo)

	roleID := uuid.New()
	permissionID := uuid.New()

	mockPermRepo.On("GetByID", mock.Anything, permissionID).Return(nil, assert.AnError)

	err := service.AssignPermissionToRole(context.Background(), roleID, permissionID)
	assert.Error(t, err)
	mockPermRepo.AssertExpectations(t)
}

