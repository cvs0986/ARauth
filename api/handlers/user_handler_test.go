package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nuage-identity/iam/api/middleware"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/identity/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of user.ServiceInterface
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Create(ctx interface{}, req *user.CreateUserRequest) (*models.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetByID(ctx interface{}, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetByUsername(ctx interface{}, username string, tenantID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, username, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetByEmail(ctx interface{}, email string, tenantID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, email, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) Update(ctx interface{}, id uuid.UUID, req *user.UpdateUserRequest) (*models.User, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) Delete(ctx interface{}, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) List(ctx interface{}, tenantID uuid.UUID, filters interface{}) ([]*models.User, error) {
	args := m.Called(ctx, tenantID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserService) Count(ctx interface{}, tenantID uuid.UUID, filters interface{}) (int64, error) {
	args := m.Called(ctx, tenantID, filters)
	return args.Get(0).(int64), args.Error(1)
}

func TestUserHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		tenantID := uuid.New()
		c.Set("tenant_id", tenantID)
		c.Next()
	})
	router.POST("/api/v1/users", handler.Create)

	reqBody := user.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	expectedUser := &models.User{
		ID:       uuid.New(),
		Username: reqBody.Username,
		Email:    reqBody.Email,
		Status:   "active",
	}

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

	router := gin.New()
	router.Use(func(c *gin.Context) {
		tenantID := uuid.New()
		c.Set("tenant_id", tenantID)
		c.Next()
	})
	router.GET("/api/v1/users/:id", handler.GetByID)

	userID := uuid.New()
	tenantID := uuid.New()
	expectedUser := &models.User{
		ID:       userID,
		TenantID: tenantID,
		Username: "testuser",
		Email:    "test@example.com",
	}

	mockService.On("GetByID", mock.Anything, userID).Return(expectedUser, nil)

	req, _ := http.NewRequest("GET", "/api/v1/users/"+userID.String(), nil)
	req.Header.Set("X-Tenant-ID", tenantID.String())
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

	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

