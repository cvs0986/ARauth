package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
)

// WebhookRepository defines the interface for webhook data access
type WebhookRepository interface {
	// Create creates a new webhook
	Create(ctx context.Context, webhook *models.Webhook) error

	// GetByID retrieves a webhook by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.Webhook, error)

	// GetByTenantID retrieves all webhooks for a tenant
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.Webhook, error)

	// GetEnabledByTenantID retrieves all enabled webhooks for a tenant
	GetEnabledByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.Webhook, error)

	// GetByEventType retrieves all enabled webhooks subscribed to an event type
	GetByEventType(ctx context.Context, tenantID uuid.UUID, eventType string) ([]*models.Webhook, error)

	// Update updates an existing webhook
	Update(ctx context.Context, webhook *models.Webhook) error

	// Delete soft deletes a webhook
	Delete(ctx context.Context, id uuid.UUID) error
}

// WebhookDeliveryRepository defines the interface for webhook delivery data access
type WebhookDeliveryRepository interface {
	// Create creates a new webhook delivery record
	Create(ctx context.Context, delivery *models.WebhookDelivery) error

	// GetByID retrieves a webhook delivery by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.WebhookDelivery, error)

	// GetByWebhookID retrieves all deliveries for a webhook
	GetByWebhookID(ctx context.Context, webhookID uuid.UUID, limit, offset int) ([]*models.WebhookDelivery, int, error)

	// GetPendingRetries retrieves all deliveries that need to be retried
	GetPendingRetries(ctx context.Context, before time.Time) ([]*models.WebhookDelivery, error)

	// Update updates a webhook delivery record
	Update(ctx context.Context, delivery *models.WebhookDelivery) error
}

