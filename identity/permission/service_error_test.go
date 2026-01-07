package permission

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Create_EmptyName(t *testing.T) {
	mockRepo := testutil.NewMockPermissionRepository()
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
	mockRepo := testutil.NewMockPermissionRepository()
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
	mockRepo := testutil.NewMockPermissionRepository()
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
	mockRepo := testutil.NewMockPermissionRepository()
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
	mockRepo := testutil.NewMockPermissionRepository()
	service := NewService(mockRepo)

	nonExistentID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, nonExistentID).Return(nil, assert.AnError)

	_, err := service.GetByID(context.Background(), nonExistentID)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

