package user

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Create_InvalidEmail(t *testing.T) {
	mockRepo := new(testutil.MockUserRepository)
	service := NewService(mockRepo)

	req := &CreateUserRequest{
		TenantID: uuid.New(),
		Username: "testuser",
		Email:    "invalid-email", // Invalid email format
		Status:   "active",
	}

	_, err := service.Create(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email")
}

func TestService_Create_EmptyUsername(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewService(mockRepo)

	req := &CreateUserRequest{
		TenantID: uuid.New(),
		Username: "", // Empty username
		Email:    "test@example.com",
		Status:   "active",
	}

	_, err := service.Create(context.Background(), req)
	assert.Error(t, err)
}

func TestService_GetByID_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewService(mockRepo)

	nonExistentID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, nonExistentID).Return(nil, assert.AnError)

	_, err := service.GetByID(context.Background(), nonExistentID)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_GetByUsername_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewService(mockRepo)

	tenantID := uuid.New()
	username := "nonexistent"
	mockRepo.On("GetByUsername", mock.Anything, username, tenantID).Return(nil, assert.AnError)

	_, err := service.GetByUsername(context.Background(), username, tenantID)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_Update_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewService(mockRepo)

	req := &UpdateUserRequest{
		Email: stringPtr("updated@example.com"),
	}

	nonExistentID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, nonExistentID).Return(nil, assert.AnError)

	_, err := service.Update(context.Background(), nonExistentID, req)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_Delete_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewService(mockRepo)

	nonExistentID := uuid.New()
	// Service directly calls repo.Delete without checking existence
	mockRepo.On("Delete", mock.Anything, nonExistentID).Return(assert.AnError)

	err := service.Delete(context.Background(), nonExistentID)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func stringPtr(s string) *string {
	return &s
}

