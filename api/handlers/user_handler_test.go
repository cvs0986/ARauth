package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/identity/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserService is a mock implementation of user service
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

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestUserHandler_Create(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	router.POST("/users", handler.Create)

	tenantID := uuid.New()
	reqBody := map[string]interface{}{
		"tenant_id": tenantID.String(),
		"username":  "testuser",
		"email":     "test@example.com",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())

	w := httptest.NewRecorder()

	expectedUser := &models.User{
		ID:       uuid.New(),
		TenantID: tenantID,
		Username: "testuser",
		Email:    "test@example.com",
	}

	mockService.On("Create", mock.Anything, mock.AnythingOfType("*user.CreateUserRequest")).Return(expectedUser, nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_GetByID(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	router.GET("/users/:id", handler.GetByID)

	userID := uuid.New()
	tenantID := uuid.New()
	expectedUser := &models.User{
		ID:       userID,
		TenantID: tenantID,
		Username: "testuser",
		Email:    "test@example.com",
	}

	req, _ := http.NewRequest("GET", "/users/"+userID.String(), nil)
	req.Header.Set("X-Tenant-ID", tenantID.String())

	w := httptest.NewRecorder()

	mockService.On("GetByID", mock.Anything, userID).Return(expectedUser, nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_Delete(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	router := setupRouter()
	router.DELETE("/users/:id", handler.Delete)

	userID := uuid.New()
	tenantID := uuid.New()

	req, _ := http.NewRequest("DELETE", "/users/"+userID.String(), nil)
	req.Header.Set("X-Tenant-ID", tenantID.String())

	w := httptest.NewRecorder()

	mockService.On("Delete", mock.Anything, userID).Return(nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

