package tenant

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/storage/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func TestService_Create(t *testing.T) {
	mockRepo := new(MockTenantRepository)
	service := NewService(mockRepo)

	req := &CreateTenantRequest{
		Name:   "Test Tenant",
		Domain: "test.example.com",
	}

	expectedTenant := &models.Tenant{
		ID:     uuid.New(),
		Name:   req.Name,
		Domain: req.Domain,
		Status: models.TenantStatusActive,
	}

	mockRepo.On("GetByDomain", mock.Anything, req.Domain).Return(nil, assert.AnError)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Tenant")).Return(nil).Run(func(args mock.Arguments) {
		tenant := args.Get(1).(*models.Tenant)
		tenant.ID = expectedTenant.ID
	})

	tenant, err := service.Create(context.Background(), req)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, tenant.ID)
	assert.Equal(t, req.Name, tenant.Name)
	assert.Equal(t, req.Domain, tenant.Domain)

	mockRepo.AssertExpectations(t)
}

func TestService_Create_DuplicateDomain(t *testing.T) {
	mockRepo := new(MockTenantRepository)
	service := NewService(mockRepo)

	req := &CreateTenantRequest{
		Name:   "Test Tenant",
		Domain: "test.example.com",
	}

	existingTenant := &models.Tenant{
		ID:     uuid.New(),
		Domain: req.Domain,
	}

	mockRepo.On("GetByDomain", mock.Anything, req.Domain).Return(existingTenant, nil)

	tenant, err := service.Create(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, tenant)
	assert.Contains(t, err.Error(), "already exists")

	mockRepo.AssertExpectations(t)
}

func TestService_GetByID(t *testing.T) {
	mockRepo := new(MockTenantRepository)
	service := NewService(mockRepo)

	tenantID := uuid.New()
	expectedTenant := &models.Tenant{
		ID:     tenantID,
		Name:   "Test Tenant",
		Domain: "test.example.com",
		Status: models.TenantStatusActive,
	}

	mockRepo.On("GetByID", mock.Anything, tenantID).Return(expectedTenant, nil)

	tenant, err := service.GetByID(context.Background(), tenantID)
	require.NoError(t, err)
	assert.Equal(t, expectedTenant.ID, tenant.ID)
	assert.Equal(t, expectedTenant.Name, tenant.Name)

	mockRepo.AssertExpectations(t)
}

func TestService_Update(t *testing.T) {
	mockRepo := new(MockTenantRepository)
	service := NewService(mockRepo)

	tenantID := uuid.New()
	existingTenant := &models.Tenant{
		ID:     tenantID,
		Name:   "Old Name",
		Domain: "old.example.com",
		Status: models.TenantStatusActive,
	}

	updatedName := "New Name"
	req := &UpdateTenantRequest{
		Name: &updatedName,
	}

	mockRepo.On("GetByID", mock.Anything, tenantID).Return(existingTenant, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Tenant")).Return(nil)

	tenant, err := service.Update(context.Background(), tenantID, req)
	require.NoError(t, err)
	assert.Equal(t, updatedName, tenant.Name)

	mockRepo.AssertExpectations(t)
}

func TestService_Delete(t *testing.T) {
	mockRepo := new(MockTenantRepository)
	service := NewService(mockRepo)

	tenantID := uuid.New()
	existingTenant := &models.Tenant{
		ID: tenantID,
	}

	mockRepo.On("GetByID", mock.Anything, tenantID).Return(existingTenant, nil)
	mockRepo.On("Delete", mock.Anything, tenantID).Return(nil)

	err := service.Delete(context.Background(), tenantID)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

