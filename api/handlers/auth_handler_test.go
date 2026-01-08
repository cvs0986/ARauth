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
	"github.com/arauth-identity/iam/auth/login"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLoginService is a mock implementation of login.ServiceInterface
type MockLoginService struct {
	mock.Mock
}

func (m *MockLoginService) Login(ctx context.Context, req *login.LoginRequest) (*login.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*login.LoginResponse), args.Error(1)
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockLoginService)
	handler := NewAuthHandler(mockService, nil) // refreshService not needed for login test

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uuid.New())
		c.Next()
	})
	router.POST("/api/v1/auth/login", handler.Login)

	reqBody := login.LoginRequest{
		Username: "testuser",
		Password: "password123",
		TenantID: uuid.New(),
	}
	body, _ := json.Marshal(reqBody)

	expectedResponse := &login.LoginResponse{
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}

	mockService.On("Login", mock.Anything, mock.AnythingOfType("*login.LoginRequest")).Return(expectedResponse, nil)

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockLoginService)
	handler := NewAuthHandler(mockService, nil) // refreshService not needed for login test

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uuid.New())
		c.Next()
	})
	router.POST("/api/v1/auth/login", handler.Login)

	// Invalid JSON
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_AuthenticationFailed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockLoginService)
	handler := NewAuthHandler(mockService, nil) // refreshService not needed for login test

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uuid.New())
		c.Next()
	})
	router.POST("/api/v1/auth/login", handler.Login)

	reqBody := login.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
		TenantID: uuid.New(),
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Login", mock.Anything, mock.AnythingOfType("*login.LoginRequest")).Return(nil, assert.AnError)

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

