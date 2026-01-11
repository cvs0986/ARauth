package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arauth-identity/iam/identity/session"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSessionService is a mock for testing
type MockSessionService struct {
	mock.Mock
}

func (m *MockSessionService) ListSessions(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]*session.Session, error) {
	args := m.Called(ctx, userID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*session.Session), args.Error(1)
}

func (m *MockSessionService) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*session.Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*session.Session), args.Error(1)
}

func (m *MockSessionService) RevokeSession(ctx context.Context, sessionID uuid.UUID, reason string) error {
	args := m.Called(ctx, sessionID, reason)
	return args.Error(0)
}

// TestListSessions_Success tests successful session listing
func TestListSessions_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSessionService := new(MockSessionService)
	mockAuditService := new(MockAuditService)
	handler := NewSessionHandler(mockSessionService, mockAuditService)

	userID := uuid.New()
	tenantID := uuid.New()

	sessions := []*session.Session{
		{
			ID:        uuid.New(),
			UserID:    userID,
			Username:  "testuser",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
	}

	mockSessionService.On("ListSessions", mock.Anything, userID, tenantID).Return(sessions, nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenantID.String())
		c.Set("user_id", userID.String())
		c.Next()
	})
	router.GET("/sessions", handler.ListSessions)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sessions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSessionService.AssertExpectations(t)
}

// TestListSessions_DeniedWithoutPermission tests that listing requires sessions:read permission
// Note: This test would require middleware setup which is tested in permission_enforcement_test.go
// Here we test the handler logic assuming permission middleware passed

// TestRevokeSession_Success tests successful session revocation
func TestRevokeSession_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSessionService := new(MockSessionService)
	mockAuditService := new(MockAuditService)
	handler := NewSessionHandler(mockSessionService, mockAuditService)

	userID := uuid.New()
	tenantID := uuid.New()
	sessionID := uuid.New()

	sessions := []*session.Session{
		{
			ID:        sessionID,
			UserID:    userID,
			Username:  "testuser",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
	}

	mockSessionService.On("ListSessions", mock.Anything, userID, tenantID).Return(sessions, nil)
	mockSessionService.On("RevokeSession", mock.Anything, sessionID, "User requested logout from web").Return(nil)
	mockAuditService.On("LogUserUpdated", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenantID.String())
		c.Set("user_id", userID.String())
		c.Set("user_claims", &struct {
			Subject       string
			Username      string
			Email         string
			PrincipalType string
		}{
			Subject:       userID.String(),
			Username:      "testuser",
			PrincipalType: "TENANT",
		})
		c.Next()
	})
	router.POST("/sessions/:id/revoke", handler.RevokeSession)

	reqBody := map[string]string{
		"audit_reason": "User requested logout from web",
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/sessions/"+sessionID.String()+"/revoke", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSessionService.AssertExpectations(t)
}

// TestRevokeSession_MissingAuditReason tests that audit_reason is required
func TestRevokeSession_MissingAuditReason(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSessionService := new(MockSessionService)
	mockAuditService := new(MockAuditService)
	handler := NewSessionHandler(mockSessionService, mockAuditService)

	userID := uuid.New()
	tenantID := uuid.New()
	sessionID := uuid.New()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenantID.String())
		c.Set("user_id", userID.String())
		c.Next()
	})
	router.POST("/sessions/:id/revoke", handler.RevokeSession)

	reqBody := map[string]string{}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/sessions/"+sessionID.String()+"/revoke", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "audit_reason")
}

// TestRevokeSession_AuditReasonTooShort tests that audit_reason must be at least 10 characters
func TestRevokeSession_AuditReasonTooShort(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSessionService := new(MockSessionService)
	mockAuditService := new(MockAuditService)
	handler := NewSessionHandler(mockSessionService, mockAuditService)

	userID := uuid.New()
	tenantID := uuid.New()
	sessionID := uuid.New()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenantID.String())
		c.Set("user_id", userID.String())
		c.Next()
	})
	router.POST("/sessions/:id/revoke", handler.RevokeSession)

	reqBody := map[string]string{
		"audit_reason": "short", // Only 5 characters
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/sessions/"+sessionID.String()+"/revoke", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRevokeSession_CrossTenantDenied tests that cross-tenant session revocation is blocked
func TestRevokeSession_CrossTenantDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSessionService := new(MockSessionService)
	mockAuditService := new(MockAuditService)
	handler := NewSessionHandler(mockSessionService, mockAuditService)

	userID := uuid.New()
	tenantID := uuid.New()
	sessionID := uuid.New()

	// Return empty sessions list (session belongs to different tenant)
	mockSessionService.On("ListSessions", mock.Anything, userID, tenantID).Return([]*session.Session{}, nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenantID.String())
		c.Set("user_id", userID.String())
		c.Next()
	})
	router.POST("/sessions/:id/revoke", handler.RevokeSession)

	reqBody := map[string]string{
		"audit_reason": "User requested logout from web",
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/sessions/"+sessionID.String()+"/revoke", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "does not belong to your tenant")
	mockSessionService.AssertExpectations(t)
}
