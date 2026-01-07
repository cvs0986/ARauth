package permission

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
	existingPermission := &models.Permission{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     permissionName,
	}
	mockRepo.On("GetByName", mock.Anything, tenantID, permissionName).Return(existingPermission, nil)

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

