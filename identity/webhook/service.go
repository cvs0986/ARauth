package webhook

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/internal/webhook"
	"github.com/arauth-identity/iam/storage/interfaces"
	"go.uber.org/zap"
)

// Service provides webhook functionality
type Service struct {
	webhookRepo    interfaces.WebhookRepository
	deliveryRepo   interfaces.WebhookDeliveryRepository
	dispatcher     *webhook.Dispatcher
	logger         *zap.Logger
}

// NewService creates a new webhook service
func NewService(
	webhookRepo interfaces.WebhookRepository,
	deliveryRepo interfaces.WebhookDeliveryRepository,
	dispatcher *webhook.Dispatcher,
	logger *zap.Logger,
) ServiceInterface {
	return &Service{
		webhookRepo:  webhookRepo,
		deliveryRepo: deliveryRepo,
		dispatcher:   dispatcher,
		logger:       logger,
	}
}

// CreateWebhook creates a new webhook
func (s *Service) CreateWebhook(ctx context.Context, tenantID uuid.UUID, req *CreateWebhookRequest) (*Webhook, error) {
	// Validate events
	if len(req.Events) == 0 {
		return nil, fmt.Errorf("at least one event type is required")
	}

	// Generate secret if not provided (should be provided, but generate as fallback)
	secret := req.Secret
	if secret == "" {
		secret = generateSecret()
	}

	w := &Webhook{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     req.Name,
		URL:      req.URL,
		Secret:   secret,
		Enabled:  req.Enabled,
		Events:   req.Events,
	}

	if err := s.webhookRepo.Create(ctx, w); err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	// Don't return secret in response
	w.Secret = ""

	return w, nil
}

// GetWebhook retrieves a webhook by ID
func (s *Service) GetWebhook(ctx context.Context, id uuid.UUID) (*Webhook, error) {
	w, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("webhook not found: %w", err)
	}

	// Don't return secret
	w.Secret = ""

	return w, nil
}

// GetWebhooksByTenant retrieves all webhooks for a tenant
func (s *Service) GetWebhooksByTenant(ctx context.Context, tenantID uuid.UUID) ([]*Webhook, error) {
	webhooks, err := s.webhookRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhooks: %w", err)
	}

	// Don't return secrets
	for _, w := range webhooks {
		w.Secret = ""
	}

	return webhooks, nil
}

// UpdateWebhook updates a webhook
func (s *Service) UpdateWebhook(ctx context.Context, id uuid.UUID, req *UpdateWebhookRequest) (*Webhook, error) {
	w, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("webhook not found: %w", err)
	}

	if req.Name != nil {
		w.Name = *req.Name
	}
	if req.URL != nil {
		w.URL = *req.URL
	}
	if req.Secret != nil {
		w.Secret = *req.Secret
	}
	if req.Enabled != nil {
		w.Enabled = *req.Enabled
	}
	if req.Events != nil {
		if len(req.Events) == 0 {
			return nil, fmt.Errorf("at least one event type is required")
		}
		w.Events = req.Events
	}

	if err := s.webhookRepo.Update(ctx, w); err != nil {
		return nil, fmt.Errorf("failed to update webhook: %w", err)
	}

	// Don't return secret
	w.Secret = ""

	return w, nil
}

// DeleteWebhook deletes a webhook
func (s *Service) DeleteWebhook(ctx context.Context, id uuid.UUID) error {
	return s.webhookRepo.Delete(ctx, id)
}

// GetDeliveriesByWebhook retrieves deliveries for a webhook
func (s *Service) GetDeliveriesByWebhook(ctx context.Context, webhookID uuid.UUID, limit, offset int) ([]*WebhookDelivery, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	return s.deliveryRepo.GetByWebhookID(ctx, webhookID, limit, offset)
}

// GetDeliveryByID retrieves a delivery by ID
func (s *Service) GetDeliveryByID(ctx context.Context, id uuid.UUID) (*WebhookDelivery, error) {
	return s.deliveryRepo.GetByID(ctx, id)
}

// TriggerWebhook triggers webhook delivery for an event
func (s *Service) TriggerWebhook(ctx context.Context, tenantID uuid.UUID, eventType string, payload map[string]interface{}, eventID *uuid.UUID) error {
	// Get all enabled webhooks subscribed to this event type
	webhooks, err := s.webhookRepo.GetByEventType(ctx, tenantID, eventType)
	if err != nil {
		return fmt.Errorf("failed to get webhooks: %w", err)
	}

	// Deliver to each webhook asynchronously
	for _, w := range webhooks {
		go func(webhook *Webhook) {
			// Create a new context for async operation
			asyncCtx := context.Background()
			if err := s.dispatcher.Deliver(asyncCtx, webhook, eventType, payload, eventID); err != nil {
				s.logger.Error("Failed to deliver webhook",
					zap.String("webhook_id", webhook.ID.String()),
					zap.String("event_type", eventType),
					zap.Error(err),
				)
			}
		}(w)
	}

	return nil
}

// generateSecret generates a random secret
func generateSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

