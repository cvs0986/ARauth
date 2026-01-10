package webhook

import (
	"context"

	"github.com/google/uuid"
)

// ServiceInterface defines the interface for webhook service operations
type ServiceInterface interface {
	// Webhook Management
	CreateWebhook(ctx context.Context, tenantID uuid.UUID, req *CreateWebhookRequest) (*Webhook, error)
	GetWebhook(ctx context.Context, id uuid.UUID) (*Webhook, error)
	GetWebhooksByTenant(ctx context.Context, tenantID uuid.UUID) ([]*Webhook, error)
	UpdateWebhook(ctx context.Context, id uuid.UUID, req *UpdateWebhookRequest) (*Webhook, error)
	DeleteWebhook(ctx context.Context, id uuid.UUID) error

	// Delivery Management
	GetDeliveriesByWebhook(ctx context.Context, webhookID uuid.UUID, limit, offset int) ([]*WebhookDelivery, int, error)
	GetDeliveryByID(ctx context.Context, id uuid.UUID) (*WebhookDelivery, error)

	// Trigger webhook delivery (called by audit service)
	TriggerWebhook(ctx context.Context, tenantID uuid.UUID, eventType string, payload map[string]interface{}, eventID *uuid.UUID) error
}

// CreateWebhookRequest represents a request to create a webhook
type CreateWebhookRequest struct {
	Name    string   `json:"name" binding:"required"`
	URL     string   `json:"url" binding:"required,url"`
	Secret  string   `json:"secret" binding:"required,min=32"` // Minimum 32 characters for security
	Enabled bool     `json:"enabled"`
	Events  []string `json:"events" binding:"required,min=1"`
}

// UpdateWebhookRequest represents a request to update a webhook
type UpdateWebhookRequest struct {
	Name    *string   `json:"name,omitempty"`
	URL     *string   `json:"url,omitempty"`
	Secret  *string   `json:"secret,omitempty"`
	Enabled *bool     `json:"enabled,omitempty"`
	Events  []string  `json:"events,omitempty"`
}

