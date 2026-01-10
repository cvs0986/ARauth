package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/identity/tenant"
	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTenantService is a mock implementation of tenant.ServiceInterface
type MockTenantService struct {
	mock.Mock
}

func (m *MockTenantService) Create(ctx context.Context, req *tenant.CreateTenantRequest) (*models.Tenant, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockTenantService) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockTenantService) GetByDomain(ctx context.Context, domain string) (*models.Tenant, error) {
	args := m.Called(ctx, domain)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockTenantService) Update(ctx context.Context, id uuid.UUID, req *tenant.UpdateTenantRequest) (*models.Tenant, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *MockTenantService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTenantService) List(ctx context.Context, filters *interfaces.TenantFilters) ([]*models.Tenant, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Tenant), args.Error(1)
}

func TestTenantHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockTenantService)
	handler := NewTenantHandler(mockService, nil)

	router := gin.New()
	router.POST("/api/v1/tenants", handler.Create)

	reqBody := tenant.CreateTenantRequest{
		Name:   "Test Tenant",
		Domain: "test.example.com",
	}
	body, _ := json.Marshal(reqBody)

	expectedTenant := &models.Tenant{
		ID:     uuid.New(),
		Name:   reqBody.Name,
		Domain: reqBody.Domain,
		Status: "active",
	}

	mockService.On("Create", mock.Anything, mock.AnythingOfType("*tenant.CreateTenantRequest")).Return(expectedTenant, nil)

	req, _ := http.NewRequest("POST", "/api/v1/tenants", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestTenantHandler_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockTenantService)
	handler := NewTenantHandler(mockService, nil)

	router := gin.New()
	router.GET("/api/v1/tenants/:id", handler.GetByID)

	tenantID := uuid.New()
	expectedTenant := &models.Tenant{
		ID:     tenantID,
		Name:   "Test Tenant",
		Domain: "test.example.com",
	}

	mockService.On("GetByID", mock.Anything, tenantID).Return(expectedTenant, nil)

	req, _ := http.NewRequest("GET", "/api/v1/tenants/"+tenantID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestTenantHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockTenantService)
	handler := NewTenantHandler(mockService, nil)

	router := gin.New()
	router.GET("/api/v1/tenants", handler.List)

	expectedTenants := []*models.Tenant{
		{ID: uuid.New(), Name: "Tenant 1"},
		{ID: uuid.New(), Name: "Tenant 2"},
	}

	mockService.On("List", mock.Anything, mock.Anything).Return(expectedTenants, nil)

	req, _ := http.NewRequest("GET", "/api/v1/tenants", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}
