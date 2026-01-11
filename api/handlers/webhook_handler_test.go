package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/identity/webhook"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWebhookService
type MockWebhookService struct {
	mock.Mock
}

func (m *MockWebhookService) CreateWebhook(ctx context.Context, tenantID uuid.UUID, req *webhook.CreateWebhookRequest) (*models.Webhook, error) {
	args := m.Called(ctx, tenantID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Webhook), args.Error(1)
}

func (m *MockWebhookService) GetWebhook(ctx context.Context, id uuid.UUID) (*models.Webhook, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Webhook), args.Error(1)
}

func (m *MockWebhookService) GetWebhooksByTenant(ctx context.Context, tenantID uuid.UUID) ([]*models.Webhook, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]*models.Webhook), args.Error(1)
}

func (m *MockWebhookService) UpdateWebhook(ctx context.Context, id uuid.UUID, req *webhook.UpdateWebhookRequest) (*models.Webhook, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Webhook), args.Error(1)
}

func (m *MockWebhookService) DeleteWebhook(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWebhookService) TriggerWebhook(ctx context.Context, tenantID uuid.UUID, eventType string, payload map[string]interface{}, eventID *uuid.UUID) error {
	args := m.Called(ctx, tenantID, eventType, payload, eventID)
	return args.Error(0)
}

func (m *MockWebhookService) GetDeliveriesByWebhook(ctx context.Context, webhookID uuid.UUID, limit, offset int) ([]*models.WebhookDelivery, int, error) {
	args := m.Called(ctx, webhookID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.WebhookDelivery), args.Int(1), args.Error(2)
}

func (m *MockWebhookService) GetDeliveryByID(ctx context.Context, id uuid.UUID) (*models.WebhookDelivery, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WebhookDelivery), args.Error(1)
}

func TestWebhookHandler_CreateWebhook(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tenantID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockService := &MockWebhookService{}
		handler := NewWebhookHandler(mockService)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("tenant_id", tenantID)
			c.Next()
		})
		router.POST("/webhooks", handler.CreateWebhook)

		reqBody := `{"name": "My Webhook", "url": "https://example.com/webhook", "secret": "12345678901234567890123456789012", "events": ["user.created"]}`

		mockService.On("CreateWebhook", mock.Anything, tenantID, mock.MatchedBy(func(req *webhook.CreateWebhookRequest) bool {
			return req.URL == "https://example.com/webhook" && req.Events[0] == "user.created" && req.Name == "My Webhook"
		})).Return(&models.Webhook{
			ID:       uuid.New(),
			TenantID: tenantID,
			URL:      "https://example.com/webhook",
			Events:   []string{"user.created"},
			Enabled:  true,
		}, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/webhooks", strings.NewReader(reqBody))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid_request", func(t *testing.T) {
		mockService := &MockWebhookService{}
		handler := NewWebhookHandler(mockService)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("tenant_id", tenantID)
			c.Next()
		})
		router.POST("/webhooks", handler.CreateWebhook)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/webhooks", strings.NewReader(`invalid json`))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestWebhookHandler_ListWebhooks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tenantID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockService := &MockWebhookService{}
		handler := NewWebhookHandler(mockService)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("tenant_id", tenantID)
			c.Next()
		})
		router.GET("/webhooks", handler.ListWebhooks)

		mockService.On("GetWebhooksByTenant", mock.Anything, tenantID).Return([]*models.Webhook{
			{ID: uuid.New(), URL: "https://example.com"},
		}, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/webhooks", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "https://example.com")
	})
}

func TestWebhookHandler_GetWebhook(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockService := &MockWebhookService{}
		handler := NewWebhookHandler(mockService)
		id := uuid.New()

		router := gin.New()
		router.GET("/webhooks/:id", handler.GetWebhook)

		mockService.On("GetWebhook", mock.Anything, id).Return(&models.Webhook{
			ID: id, URL: "https://example.com",
		}, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/webhooks/"+id.String(), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("not_found", func(t *testing.T) {
		mockService := &MockWebhookService{}
		handler := NewWebhookHandler(mockService)
		id := uuid.New()

		router := gin.New()
		router.GET("/webhooks/:id", handler.GetWebhook)

		mockService.On("GetWebhook", mock.Anything, id).Return(nil, errors.New("not found"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/webhooks/"+id.String(), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestWebhookHandler_DeleteWebhook(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockService := &MockWebhookService{}
		handler := NewWebhookHandler(mockService)

		router := gin.New()
		router.DELETE("/webhooks/:id", handler.DeleteWebhook)

		mockService.On("DeleteWebhook", mock.Anything, id).Return(nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/webhooks/"+id.String(), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}
