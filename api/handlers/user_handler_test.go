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
	"github.com/nuage-identity/iam/identity/user"
	"github.com/nuage-identity/iam/storage/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of user.ServiceInterface
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Create(ctx context.Context, req *user.CreateUserRequest) (*models.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetByUsername(ctx context.Context, username string, tenantID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, username, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetByEmail(ctx context.Context, email string, tenantID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, email, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) Update(ctx context.Context, id uuid.UUID, req *user.UpdateUserRequest) (*models.User, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) ([]*models.User, error) {
	args := m.Called(ctx, tenantID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserService) Count(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) (int, error) {
	args := m.Called(ctx, tenantID, filters)
	return args.Get(0).(int), args.Error(1)
}

func TestUserHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	tenantID := uuid.New()
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenantID)
		c.Next()
	})
	router.POST("/api/v1/users", handler.Create)

	reqBody := user.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
	}
	body, _ := json.Marshal(reqBody)

	expectedUser := &models.User{
		ID:       uuid.New(),
		TenantID: tenantID,
		Username: reqBody.Username,
		Email:    reqBody.Email,
		Status:   "active",
	}

	// The handler sets tenant_id from context, so we need to match that
	mockService.On("Create", mock.Anything, mock.AnythingOfType("*user.CreateUserRequest")).Return(expectedUser, nil)

	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	userID := uuid.New()
	tenantID := uuid.New()
	expectedUser := &models.User{
		ID:       userID,
		TenantID: tenantID,
		Username: "testuser",
		Email:    "test@example.com",
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Set tenant ID in context to match user's tenant ID
		c.Set("tenant_id", tenantID)
		c.Next()
	})
	router.GET("/api/v1/users/:id", handler.GetByID)

	mockService.On("GetByID", mock.Anything, userID).Return(expectedUser, nil)

	req, _ := http.NewRequest("GET", "/api/v1/users/"+userID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		tenantID := uuid.New()
		c.Set("tenant_id", tenantID)
		c.Next()
	})
	router.GET("/api/v1/users", handler.List)

	expectedUsers := []*models.User{
		{ID: uuid.New(), Username: "user1"},
		{ID: uuid.New(), Username: "user2"},
	}

	mockService.On("List", mock.Anything, mock.Anything, mock.Anything).Return(expectedUsers, nil)
	mockService.On("Count", mock.Anything, mock.Anything, mock.Anything).Return(len(expectedUsers), nil)

	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

