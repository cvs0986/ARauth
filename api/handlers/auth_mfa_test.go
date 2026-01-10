package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arauth-identity/iam/auth/login"
	"github.com/arauth-identity/iam/auth/mfa"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthMFAService is a mock implementation of mfa.ServiceInterface
type MockAuthMFAService struct {
	mock.Mock
}

func (m *MockAuthMFAService) Enroll(ctx context.Context, req *mfa.EnrollRequest) (*mfa.EnrollResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mfa.EnrollResponse), args.Error(1)
}

func (m *MockAuthMFAService) EnrollForLogin(ctx context.Context, sessionID string) (*mfa.EnrollResponse, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mfa.EnrollResponse), args.Error(1)
}

func (m *MockAuthMFAService) Verify(ctx context.Context, req *mfa.VerifyRequest) (bool, error) {
	args := m.Called(ctx, req)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthMFAService) CreateChallenge(ctx context.Context, req *mfa.ChallengeRequest) (*mfa.ChallengeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mfa.ChallengeResponse), args.Error(1)
}

func (m *MockAuthMFAService) VerifyChallenge(ctx context.Context, req *mfa.VerifyChallengeRequest) (*mfa.VerifyChallengeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mfa.VerifyChallengeResponse), args.Error(1)
}

func (m *MockAuthMFAService) CreateSession(ctx context.Context, userID, tenantID uuid.UUID) (string, error) {
	args := m.Called(ctx, userID, tenantID)
	return args.String(0), args.Error(1)
}

// MockAuthAuditService satisfies audit.ServiceInterface
type MockAuthAuditService struct {
	mock.Mock
}

func (m *MockAuthAuditService) LogEvent(ctx context.Context, event *models.AuditEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}
func (m *MockAuthAuditService) QueryEvents(ctx context.Context, filters *interfaces.AuditEventFilters) ([]*models.AuditEvent, int, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*models.AuditEvent), args.Int(1), args.Error(2)
}
func (m *MockAuthAuditService) GetEvent(ctx context.Context, eventID uuid.UUID) (*models.AuditEvent, error) {
	args := m.Called(ctx, eventID)
	return args.Get(0).(*models.AuditEvent), args.Error(1)
}
func (m *MockAuthAuditService) LogUserCreated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error {
	return nil
}
func (m *MockAuthAuditService) LogUserUpdated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error {
	return nil
}
func (m *MockAuthAuditService) LogUserDeleted(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogUserLocked(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogUserUnlocked(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogRoleAssigned(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error {
	return nil
}
func (m *MockAuthAuditService) LogRoleRemoved(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogRoleCreated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogRoleUpdated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogRoleDeleted(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogPermissionAssigned(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogPermissionRemoved(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogPermissionCreated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogPermissionUpdated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogPermissionDeleted(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogMFAEnrolled(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogMFAVerified(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string, success bool) error {
	return nil
}
func (m *MockAuthAuditService) LogMFADisabled(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogMFAReset(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogTenantCreated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogTenantUpdated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogTenantDeleted(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogTenantSuspended(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogTenantResumed(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogTenantSettingsUpdated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, sourceIP, userAgent string) error {
	return nil
}
func (m *MockAuthAuditService) LogLoginSuccess(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error {
	return nil
}
func (m *MockAuthAuditService) LogLoginFailure(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string, reason string) error {
	return nil
}
func (m *MockAuthAuditService) LogTokenIssued(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error {
	return nil
}
func (m *MockAuthAuditService) LogTokenRevoked(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error {
	return nil
}

// NEW Method
func (m *MockAuthAuditService) LogMFAChallengeCreated(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	args := m.Called(ctx, actor, tenantID, sourceIP, userAgent)
	return args.Error(0)
}

func TestAuthHandler_Login_MFARequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLoginService := new(MockLoginService)
	mockMFAService := new(MockAuthMFAService)
	mockAuditService := new(MockAuthAuditService)

	// Since we changed NewAuthHandler signature, we must pass mfaService
	handler := NewAuthHandler(mockLoginService, nil, nil, mockAuditService, mockMFAService)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", uuid.New())
		c.Next() // IMPORTANT: gin.New() doesn't have Logger/Recovery, but tests usually don't need them.
	})
	router.POST("/api/v1/auth/login", handler.Login)

	reqBody := login.LoginRequest{
		Username: "mfauser",
		Password: "password123",
		TenantID: uuid.New(),
	}
	// We expect request body binding to succeed and Login service to be called

	// SETUP: MFARequired=true response
	userID := uuid.New()
	tenantID := reqBody.TenantID

	expectedLoginResponse := &login.LoginResponse{
		MFARequired: true,
		UserID:      userID.String(),
		TenantID:    tenantID.String(),
	}

	// MOCK: LoginService returns MFARequired
	mockLoginService.On("Login", mock.Anything, mock.AnythingOfType("*login.LoginRequest")).Return(expectedLoginResponse, nil)

	// MOCK: MFAService.CreateSession is called
	expectedSessionID := "mfa-session-123"
	mockMFAService.On("CreateSession", mock.Anything, userID, tenantID).Return(expectedSessionID, nil)

	// MOCK: AuditService.LogMFAChallengeCreated is called
	mockAuditService.On("LogMFAChallengeCreated", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// EXECUTE
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// ASSERT
	assert.Equal(t, http.StatusOK, w.Code)

	var resp login.LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.True(t, resp.MFARequired)
	assert.Equal(t, expectedSessionID, resp.MFASessionID)
	assert.Empty(t, resp.AccessToken) // CRITICAL: No tokens

	mockLoginService.AssertExpectations(t)
	mockMFAService.AssertExpectations(t)
	mockAuditService.AssertExpectations(t)
}
