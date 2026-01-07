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
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/identity/role"
	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRoleService is a mock implementation of role.ServiceInterface
type MockRoleService struct {
	mock.Mock
}

func (m *MockRoleService) Create(ctx context.Context, req *role.CreateRoleRequest) (*models.Role, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleService) GetByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleService) Update(ctx context.Context, id uuid.UUID, req *role.UpdateRoleRequest) (*models.Role, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoleService) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.RoleFilters) ([]*models.Role, error) {
	args := m.Called(ctx, tenantID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Role), args.Error(1)
}

func (m *MockRoleService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.Role, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Role), args.Error(1)
}

func (m *MockRoleService) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockRoleService) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockRoleService) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Permission), args.Error(1)
}

func (m *MockRoleService) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func (m *MockRoleService) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func TestRoleHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockRoleService)
	handler := NewRoleHandler(mockService)

	tenantID := uuid.New()
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenantID)
		c.Next()
	})
	router.POST("/api/v1/roles", handler.Create)

	reqBody := role.CreateRoleRequest{
		TenantID: tenantID,
		Name:     "Admin",
	}
	body, _ := json.Marshal(reqBody)

	expectedRole := &models.Role{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     reqBody.Name,
	}

	mockService.On("Create", mock.Anything, mock.AnythingOfType("*role.CreateRoleRequest")).Return(expectedRole, nil)

	req, _ := http.NewRequest("POST", "/api/v1/roles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestRoleHandler_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockRoleService)
	handler := NewRoleHandler(mockService)

	tenantID := uuid.New()
	roleID := uuid.New()
	expectedRole := &models.Role{
		ID:       roleID,
		TenantID: tenantID,
		Name:     "Admin",
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenantID)
		c.Next()
	})
	router.GET("/api/v1/roles/:id", handler.GetByID)

	mockService.On("GetByID", mock.Anything, roleID).Return(expectedRole, nil)

	req, _ := http.NewRequest("GET", "/api/v1/roles/"+roleID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestRoleHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockRoleService)
	handler := NewRoleHandler(mockService)

	tenantID := uuid.New()
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenantID)
		c.Next()
	})
	router.GET("/api/v1/roles", handler.List)

	expectedRoles := []*models.Role{
		{ID: uuid.New(), Name: "Admin"},
		{ID: uuid.New(), Name: "User"},
	}

	mockService.On("List", mock.Anything, tenantID, mock.Anything).Return(expectedRoles, nil)

	req, _ := http.NewRequest("GET", "/api/v1/roles", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

