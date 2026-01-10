package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/webhook"
)

// WebhookRepository defines the interface for webhook data access
type WebhookRepository interface {
	// Create creates a new webhook
	Create(ctx context.Context, webhook *webhook.Webhook) error

	// GetByID retrieves a webhook by ID
	GetByID(ctx context.Context, id uuid.UUID) (*webhook.Webhook, error)

	// GetByTenantID retrieves all webhooks for a tenant
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*webhook.Webhook, error)

	// GetEnabledByTenantID retrieves all enabled webhooks for a tenant
	GetEnabledByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*webhook.Webhook, error)

	// GetByEventType retrieves all enabled webhooks subscribed to an event type
	GetByEventType(ctx context.Context, tenantID uuid.UUID, eventType string) ([]*webhook.Webhook, error)

	// Update updates an existing webhook
	Update(ctx context.Context, webhook *webhook.Webhook) error

	// Delete soft deletes a webhook
	Delete(ctx context.Context, id uuid.UUID) error
}

// WebhookDeliveryRepository defines the interface for webhook delivery data access
type WebhookDeliveryRepository interface {
	// Create creates a new webhook delivery record
	Create(ctx context.Context, delivery *webhook.WebhookDelivery) error

	// GetByID retrieves a webhook delivery by ID
	GetByID(ctx context.Context, id uuid.UUID) (*webhook.WebhookDelivery, error)

	// GetByWebhookID retrieves all deliveries for a webhook
	GetByWebhookID(ctx context.Context, webhookID uuid.UUID, limit, offset int) ([]*webhook.WebhookDelivery, int, error)

	// GetPendingRetries retrieves all deliveries that need to be retried
	GetPendingRetries(ctx context.Context, before time.Time) ([]*webhook.WebhookDelivery, error)

	// Update updates a webhook delivery record
	Update(ctx context.Context, delivery *webhook.WebhookDelivery) error
}

