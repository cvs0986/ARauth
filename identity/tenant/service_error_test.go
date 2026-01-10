package tenant

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Create_EmptyName(t *testing.T) {
	mockRepo := new(MockTenantRepository)
	service := NewService(mockRepo, nil)

	req := &CreateTenantRequest{
		Name:   "", // Empty name - service doesn't validate this, only domain
		Domain: "test.example.com",
	}

	// Service only validates domain, not name
	// Name validation may be at handler level via binding tags
	mockRepo.On("GetByDomain", mock.Anything, "test.example.com").Return(nil, assert.AnError)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Tenant")).Return(nil)

	_, err := service.Create(context.Background(), req)
	// Service allows empty name - validation is at handler level
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_Create_EmptyDomain(t *testing.T) {
	mockRepo := new(MockTenantRepository)
	service := NewService(mockRepo, nil)

	req := &CreateTenantRequest{
		Name:   "Test Tenant",
		Domain: "", // Empty domain
	}

	_, err := service.Create(context.Background(), req)
	assert.Error(t, err)
}

func TestService_Create_InvalidDomain(t *testing.T) {
	mockRepo := new(MockTenantRepository)
	service := NewService(mockRepo, nil)

	req := &CreateTenantRequest{
		Name:   "Test Tenant",
		Domain: "invalid-domain", // Invalid domain format
	}

	_, err := service.Create(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "domain")
}

func TestService_Create_DuplicateDomain_Error(t *testing.T) {
	mockRepo := new(MockTenantRepository)
	service := NewService(mockRepo, nil)

	domain := "test.example.com"

	// Mock existing tenant
	existingTenant := &models.Tenant{
		ID:     uuid.New(),
		Domain: domain,
	}
	mockRepo.On("GetByDomain", mock.Anything, domain).Return(existingTenant, nil)

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
	mockRepo := new(MockTenantRepository)
	service := NewService(mockRepo, nil)

	nonExistentID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, nonExistentID).Return(nil, assert.AnError)

	_, err := service.GetByID(context.Background(), nonExistentID)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_GetByDomain_NotFound(t *testing.T) {
	mockRepo := new(MockTenantRepository)
	service := NewService(mockRepo, nil)

	domain := "nonexistent.example.com"
	mockRepo.On("GetByDomain", mock.Anything, domain).Return(nil, assert.AnError)

	_, err := service.GetByDomain(context.Background(), domain)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

