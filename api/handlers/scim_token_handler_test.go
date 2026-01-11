package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arauth-identity/iam/identity/audit"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/identity/scim"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// MockSCIMTokenService
type MockSCIMTokenService struct {
	scim.TokenServiceInterface
	RotateTokenFunc func(ctx context.Context, id uuid.UUID) (*models.SCIMToken, string, error)
	CreateTokenFunc func(ctx context.Context, tenantID uuid.UUID, req *scim.CreateTokenRequest) (*models.SCIMToken, string, error)
	GetTokenFunc    func(ctx context.Context, id uuid.UUID) (*models.SCIMToken, error)
	DeleteTokenFunc func(ctx context.Context, id uuid.UUID) error
	ListTokensFunc  func(ctx context.Context, tenantID uuid.UUID) ([]*models.SCIMToken, error)
}

func (m *MockSCIMTokenService) RotateToken(ctx context.Context, id uuid.UUID) (*models.SCIMToken, string, error) {
	if m.RotateTokenFunc != nil {
		return m.RotateTokenFunc(ctx, id)
	}
	return nil, "", nil
}

func (m *MockSCIMTokenService) CreateToken(ctx context.Context, tenantID uuid.UUID, req *scim.CreateTokenRequest) (*models.SCIMToken, string, error) {
	if m.CreateTokenFunc != nil {
		return m.CreateTokenFunc(ctx, tenantID, req)
	}
	return nil, "", nil
}

func (m *MockSCIMTokenService) GetToken(ctx context.Context, id uuid.UUID) (*models.SCIMToken, error) {
	if m.GetTokenFunc != nil {
		return m.GetTokenFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSCIMTokenService) DeleteToken(ctx context.Context, id uuid.UUID) error {
	if m.DeleteTokenFunc != nil {
		return m.DeleteTokenFunc(ctx, id)
	}
	return nil
}

func (m *MockSCIMTokenService) ListTokens(ctx context.Context, tenantID uuid.UUID) ([]*models.SCIMToken, error) {
	if m.ListTokensFunc != nil {
		return m.ListTokensFunc(ctx, tenantID)
	}
	return nil, nil
}

// MockSCIMTokenAuditService
type MockSCIMTokenAuditService struct {
	audit.ServiceInterface
	LogTokenIssuedFunc  func(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error
	LogTokenRevokedFunc func(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error
}

func (m *MockSCIMTokenAuditService) LogTokenIssued(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error {
	if m.LogTokenIssuedFunc != nil {
		return m.LogTokenIssuedFunc(ctx, actor, tenantID, sourceIP, userAgent, metadata)
	}
	return nil
}

func (m *MockSCIMTokenAuditService) LogTokenRevoked(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error {
	if m.LogTokenRevokedFunc != nil {
		return m.LogTokenRevokedFunc(ctx, actor, tenantID, sourceIP, userAgent, metadata)
	}
	return nil
}

func TestSCIMTokenHandler_RotateToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tenantID := uuid.New()
	tokenID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockService := &MockSCIMTokenService{}
		mockAudit := &MockSCIMTokenAuditService{}
		handler := NewSCIMTokenHandler(mockService, mockAudit)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("tenant_id", tenantID)
			c.Set("user_id", uuid.New().String()) // For actor extraction if needed
			c.Set("username", "admin")
			c.Set("principal_type", "TENANT")
			c.Next()
		})
		router.POST("/scim/tokens/:id/rotate", handler.RotateToken)

		mockService.RotateTokenFunc = func(ctx context.Context, id uuid.UUID) (*models.SCIMToken, string, error) {
			assert.Equal(t, tokenID, id)
			return &models.SCIMToken{
				ID:       tokenID,
				TenantID: tenantID,
				Name:     "Test Token",
			}, "new-secret", nil
		}

		mockAudit.LogTokenIssuedFunc = func(ctx context.Context, actor models.AuditActor, tid *uuid.UUID, ip, ua string, meta map[string]interface{}) error {
			assert.Equal(t, tenantID, *tid)
			assert.Equal(t, "rotation", meta["action"])
			return nil
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/scim/tokens/"+tokenID.String()+"/rotate", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp gin.H
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "new-secret", resp["plaintext_token"])
	})

	t.Run("not_found", func(t *testing.T) {
		mockService := &MockSCIMTokenService{}
		handler := NewSCIMTokenHandler(mockService, &MockSCIMTokenAuditService{})

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("tenant_id", tenantID)
			c.Next()
		})
		router.POST("/scim/tokens/:id/rotate", handler.RotateToken)

		mockService.RotateTokenFunc = func(ctx context.Context, id uuid.UUID) (*models.SCIMToken, string, error) {
			return nil, "", errors.New("token not found")
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/scim/tokens/"+tokenID.String()+"/rotate", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		mockService := &MockSCIMTokenService{}
		handler := NewSCIMTokenHandler(mockService, &MockSCIMTokenAuditService{})

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("tenant_id", tenantID)
			c.Next()
		})
		router.POST("/scim/tokens/:id/rotate", handler.RotateToken)

		mockService.RotateTokenFunc = func(ctx context.Context, id uuid.UUID) (*models.SCIMToken, string, error) {
			return &models.SCIMToken{
				ID:       tokenID,
				TenantID: uuid.New(), // Different tenant
				Name:     "Test Token",
			}, "new-secret", nil
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/scim/tokens/"+tokenID.String()+"/rotate", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
