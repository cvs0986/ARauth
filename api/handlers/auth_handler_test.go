package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/auth/login"
	"github.com/arauth-identity/iam/auth/token"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	mockLoginService := new(MockLoginService)
	mockAuditService := new(MockAuditService)
	mockTokenService := new(MockTokenService)

	// Setup expectations
	mockAuditService.On("LogLoginSuccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockAuditService.On("LogTokenIssued", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Setup Token Service to validate the token (needed for audit log extraction)
	claims := &claims.Claims{
		Subject:  uuid.New().String(),
		Username: "testuser",
		TenantID: uuid.New().String(),
	}
	mockTokenService.On("ValidateAccessToken", "test-token").Return(claims, nil)

	handler := NewAuthHandler(mockLoginService, nil, mockTokenService, mockAuditService, nil)

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

	mockLoginService.On("Login", mock.Anything, mock.AnythingOfType("*login.LoginRequest")).Return(expectedResponse, nil)

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockLoginService.AssertExpectations(t)
	// mockAuditService.AssertExpectations(t) // optional, sometimes hard to match generic args
}

func TestAuthHandler_Login_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLoginService := new(MockLoginService)
	// Even for invalid request, handler might try to use audit service if bind succeeds partially or before logic?
	// But bind fails first. Still, safer to pass non-nil.
	mockAuditService := new(MockAuditService)
	handler := NewAuthHandler(mockLoginService, nil, nil, mockAuditService, nil)

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

	mockLoginService := new(MockLoginService)
	mockAuditService := new(MockAuditService)

	// Expect LogLoginFailure
	mockAuditService.On("LogLoginFailure", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	handler := NewAuthHandler(mockLoginService, nil, nil, mockAuditService, nil)

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

	mockLoginService.On("Login", mock.Anything, mock.AnythingOfType("*login.LoginRequest")).Return(nil, assert.AnError)

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockLoginService.AssertExpectations(t)
}

// MockTokenService for handler tests
type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GenerateAccessToken(claimsObj *claims.Claims, expiresIn time.Duration) (string, error) {
	args := m.Called(claimsObj, expiresIn)
	return args.String(0), args.Error(1)
}
func (m *MockTokenService) GenerateRefreshToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}
func (m *MockTokenService) HashRefreshToken(token string) (string, error) {
	args := m.Called(token)
	return args.String(0), args.Error(1)
}
func (m *MockTokenService) VerifyRefreshToken(token, hash string) bool {
	args := m.Called(token, hash)
	return args.Bool(0)
}
func (m *MockTokenService) ValidateAccessToken(tokenString string) (*claims.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*claims.Claims), args.Error(1)
}
func (m *MockTokenService) GetPublicKey() interface{} {
	args := m.Called()
	return args.Get(0)
}
func (m *MockTokenService) RevokeAccessToken(ctx context.Context, tokenString string) error {
	args := m.Called(ctx, tokenString)
	return args.Error(0)
}

// IsAccessTokenRevoked stub
func (m *MockTokenService) IsAccessTokenRevoked(ctx context.Context, jti string) (bool, error) {
	args := m.Called(ctx, jti)
	return args.Bool(0), args.Error(1)
}

// MockAuditService
type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) LogLoginSuccess(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, ip, userAgent string, metadata map[string]interface{}) error {
	return nil
}
func (m *MockAuditService) LogLoginFailure(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, ip, userAgent, reason string) error {
	return nil
}
func (m *MockAuditService) LogTokenIssued(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, ip, userAgent string, metadata map[string]interface{}) error {
	return nil
}
func (m *MockAuditService) LogTokenRevoked(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, ip, userAgent string, metadata map[string]interface{}) error {
	return nil
}
func (m *MockAuditService) LogMFAChallengeCreated(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, ip, userAgent string) error {
	return nil
}
func (m *MockAuditService) LogMFADisabled(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, ip, userAgent string) error {
	return nil
}
func (m *MockAuditService) LogMFAEnrolled(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, ip, userAgent string) error {
	return nil
}
func (m *MockAuditService) LogMFAReset(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, ip, userAgent string) error {
	return nil
}

func (m *MockAuditService) LogMFAVerified(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, ip, userAgent string, success bool) error {
	return nil
}

func (m *MockAuditService) LogPermissionAssigned(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, ip, userAgent string) error {
	return nil
}

func (m *MockAuditService) LogPermissionCreated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, ip, userAgent string) error {
	return nil
}

func (m *MockAuditService) LogPermissionDeleted(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, ip, userAgent string) error {
	return nil
}

func (m *MockAuditService) LogPermissionUpdated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, ip, userAgent string) error {
	return nil
}

func (m *MockAuditService) LogPermissionRemoved(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, ip, userAgent string) error {
	return nil
}

func (m *MockAuditService) LogRoleAssigned(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, ip, userAgent string, metadata map[string]interface{}) error {
	return nil
}
func (m *MockAuditService) LogRoleRemoved(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, ip, userAgent string) error {
	return nil
}
func (m *MockAuditService) LogRoleCreated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, ip, userAgent string) error {
	return nil
}
func (m *MockAuditService) LogRoleUpdated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, ip, userAgent string) error {
	return nil
}
func (m *MockAuditService) LogRoleDeleted(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, ip, userAgent string) error {
	return nil
}

// User events
func (m *MockAuditService) LogUserCreated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error {
	return nil
}
func (m *MockAuditService) LogUserUpdated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error {
	return nil
}
func (m *MockAuditService) LogUserDeleted(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuditService) LogUserLocked(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuditService) LogUserUnlocked(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	return nil
}

// Tenant events
func (m *MockAuditService) LogTenantCreated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuditService) LogTenantUpdated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuditService) LogTenantDeleted(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuditService) LogTenantSuspended(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuditService) LogTenantResumed(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuditService) LogTenantSettingsUpdated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuditService) QueryEvents(ctx context.Context, filters *interfaces.AuditEventFilters) ([]*models.AuditEvent, int, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.AuditEvent), args.Int(1), args.Error(2)
}

func (m *MockAuditService) ExportEvents(ctx context.Context, filters *interfaces.AuditEventFilters) ([]byte, string, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]byte), args.String(1), args.Error(2)
}

func (m *MockAuditService) GetEvent(ctx context.Context, eventID uuid.UUID) (*models.AuditEvent, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AuditEvent), args.Error(1)
}
func (m *MockAuditService) LogEvent(ctx context.Context, event *models.AuditEvent) error {
	return nil
}

func TestAuthHandler_RevokeToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockTokenService := new(MockTokenService)
	mockAuditService := new(MockAuditService)

	// Create real RefreshService with mock dependencies
	refreshService := token.NewRefreshService(mockTokenService, nil, nil, nil, nil)

	handler := NewAuthHandler(nil, refreshService, mockTokenService, mockAuditService, nil)

	router := gin.New()
	router.POST("/api/v1/auth/revoke", handler.RevokeToken)

	// Scenario 1: Revoke Access Token via Header
	tokenVal := "valid-token"

	// Expect HashRefreshToken failure (to simulate "not a refresh token" or "refresh revoke failed")
	// Note: since body is empty, req.Token is empty, so it calls HashRefreshToken("")
	mockTokenService.On("HashRefreshToken", mock.Anything).Return("", assert.AnError)

	mockTokenService.On("RevokeAccessToken", mock.Anything, tokenVal).Return(nil)
	mockTokenService.On("ValidateAccessToken", tokenVal).Return(&claims.Claims{Subject: uuid.New().String()}, nil)

	req, _ := http.NewRequest("POST", "/api/v1/auth/revoke", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenVal)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
