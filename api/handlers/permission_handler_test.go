package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/identity/permission"
	"github.com/nuage-identity/iam/storage/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPermissionService is a mock implementation of permission.ServiceInterface
type MockPermissionService struct {
	mock.Mock
}

func (m *MockPermissionService) Create(ctx context.Context, req *permission.CreatePermissionRequest) (*models.Permission, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Permission), args.Error(1)
}

func (m *MockPermissionService) GetByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Permission), args.Error(1)
}

func (m *MockPermissionService) Update(ctx context.Context, id uuid.UUID, req *permission.UpdatePermissionRequest) (*models.Permission, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Permission), args.Error(1)
}

func (m *MockPermissionService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPermissionService) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.PermissionFilters) ([]*models.Permission, error) {
	args := m.Called(ctx, tenantID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Permission), args.Error(1)
}

func TestPermissionHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockPermissionService)
	handler := NewPermissionHandler(mockService)

	tenantID := uuid.New()
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenantID)
		c.Next()
	})
	router.POST("/api/v1/permissions", handler.Create)

	reqBody := permission.CreatePermissionRequest{
		TenantID: tenantID,
		Name:     "users:read",
		Resource: "users",
		Action:   "read",
	}
	body, _ := json.Marshal(reqBody)

	expectedPermission := &models.Permission{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     reqBody.Name,
		Resource: reqBody.Resource,
		Action:   reqBody.Action,
	}

	mockService.On("Create", mock.Anything, mock.AnythingOfType("*permission.CreatePermissionRequest")).Return(expectedPermission, nil)

	req, _ := http.NewRequest("POST", "/api/v1/permissions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestPermissionHandler_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockPermissionService)
	handler := NewPermissionHandler(mockService)

	tenantID := uuid.New()
	permissionID := uuid.New()
	expectedPermission := &models.Permission{
		ID:       permissionID,
		TenantID: tenantID,
		Name:     "users:read",
		Resource: "users",
		Action:   "read",
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenantID)
		c.Next()
	})
	router.GET("/api/v1/permissions/:id", handler.GetByID)

	mockService.On("GetByID", mock.Anything, permissionID).Return(expectedPermission, nil)

	req, _ := http.NewRequest("GET", "/api/v1/permissions/"+permissionID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestPermissionHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockPermissionService)
	handler := NewPermissionHandler(mockService)

	tenantID := uuid.New()
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenantID)
		c.Next()
	})
	router.GET("/api/v1/permissions", handler.List)

	expectedPermissions := []*models.Permission{
		{ID: uuid.New(), Name: "users:read"},
		{ID: uuid.New(), Name: "users:write"},
	}

	mockService.On("List", mock.Anything, tenantID, mock.Anything).Return(expectedPermissions, nil)

	req, _ := http.NewRequest("GET", "/api/v1/permissions", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

