package tenant

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Create_EmptyName(t *testing.T) {
	mockRepo := testutil.NewMockTenantRepository()
	service := NewService(mockRepo)

	req := &CreateTenantRequest{
		Name:   "", // Empty name
		Domain: "test.example.com",
	}

	_, err := service.Create(context.Background(), req)
	assert.Error(t, err)
}

func TestService_Create_EmptyDomain(t *testing.T) {
	mockRepo := testutil.NewMockTenantRepository()
	service := NewService(mockRepo)

	req := &CreateTenantRequest{
		Name:   "Test Tenant",
		Domain: "", // Empty domain
	}

	_, err := service.Create(context.Background(), req)
	assert.Error(t, err)
}

func TestService_Create_InvalidDomain(t *testing.T) {
	mockRepo := testutil.NewMockTenantRepository()
	service := NewService(mockRepo)

	req := &CreateTenantRequest{
		Name:   "Test Tenant",
		Domain: "invalid-domain", // Invalid domain format
	}

	_, err := service.Create(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "domain")
}

func TestService_Create_DuplicateDomain(t *testing.T) {
	mockRepo := testutil.NewMockTenantRepository()
	service := NewService(mockRepo)

	domain := "test.example.com"

	// Mock existing tenant
	mockRepo.On("GetByDomain", mock.Anything, domain).Return(&struct {
		ID     uuid.UUID
		Domain string
	}{ID: uuid.New(), Domain: domain}, nil)

	req := &CreateTenantRequest{
		Name:   "Test Tenant",
		Domain: domain,
	}

	_, err := service.Create(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	mockRepo.AssertExpectations(t)
}

func TestService_GetByID_NotFound(t *testing.T) {
	mockRepo := testutil.NewMockTenantRepository()
	service := NewService(mockRepo)

	nonExistentID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, nonExistentID).Return(nil, assert.AnError)

	_, err := service.GetByID(context.Background(), nonExistentID)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_GetByDomain_NotFound(t *testing.T) {
	mockRepo := testutil.NewMockTenantRepository()
	service := NewService(mockRepo)

	domain := "nonexistent.example.com"
	mockRepo.On("GetByDomain", mock.Anything, domain).Return(nil, assert.AnError)

	_, err := service.GetByDomain(context.Background(), domain)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

