package tenant

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/storage/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTenantRepository is a mock implementation of TenantRepository
type MockTenantRepository struct {
	mock.Mock
}

func (m *MockTenantRepository) Create(ctx context.Context, t *models.Tenant) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *MockTenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockTenantRepository) GetByDomain(ctx context.Context, domain string) (*models.Tenant, error) {
	args := m.Called(ctx, domain)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockTenantRepository) Update(ctx context.Context, t *models.Tenant) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *MockTenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTenantRepository) List(ctx context.Context, filters *interfaces.TenantFilters) ([]*models.Tenant, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Tenant), args.Error(1)
}

func TestService_Create_EmptyName(t *testing.T) {
	mockRepo := new(MockTenantRepository)
	service := NewService(mockRepo)

	req := &CreateTenantRequest{
		Name:   "", // Empty name
		Domain: "test.example.com",
	}

	_, err := service.Create(context.Background(), req)
	assert.Error(t, err)
}

func TestService_Create_EmptyDomain(t *testing.T) {
	mockRepo := new(MockTenantRepository)
	service := NewService(mockRepo)

	req := &CreateTenantRequest{
		Name:   "Test Tenant",
		Domain: "", // Empty domain
	}

	_, err := service.Create(context.Background(), req)
	assert.Error(t, err)
}

func TestService_Create_InvalidDomain(t *testing.T) {
	mockRepo := new(MockTenantRepository)
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
	mockRepo := new(MockTenantRepository)
	service := NewService(mockRepo)

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
	service := NewService(mockRepo)

	nonExistentID := uuid.New()
	mockRepo.On("GetByID", mock.Anything, nonExistentID).Return(nil, assert.AnError)

	_, err := service.GetByID(context.Background(), nonExistentID)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_GetByDomain_NotFound(t *testing.T) {
	mockRepo := new(MockTenantRepository)
	service := NewService(mockRepo)

	domain := "nonexistent.example.com"
	mockRepo.On("GetByDomain", mock.Anything, domain).Return(nil, assert.AnError)

	_, err := service.GetByDomain(context.Background(), domain)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

