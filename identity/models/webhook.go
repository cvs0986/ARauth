package models

import (
	"time"

	"github.com/google/uuid"
)

// Webhook represents a webhook configuration
type Webhook struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TenantID  uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Name      string    `json:"name" db:"name"`
	URL       string    `json:"url" db:"url"`
	Secret    string    `json:"secret" db:"secret"` // Note: Should not be returned in API responses
	Enabled   bool      `json:"enabled" db:"enabled"`
	Events    []string  `json:"events" db:"events"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	WebhookID     uuid.UUID              `json:"webhook_id" db:"webhook_id"`
	EventID       *uuid.UUID             `json:"event_id,omitempty" db:"event_id"`
	EventType     string                 `json:"event_type" db:"event_type"`
	Payload       map[string]interface{} `json:"payload" db:"payload"`
	Status        DeliveryStatus         `json:"status" db:"status"`
	HTTPStatusCode *int                  `json:"http_status_code,omitempty" db:"http_status_code"`
	ResponseBody  *string                `json:"response_body,omitempty" db:"response_body"`
	AttemptNumber int                    `json:"attempt_number" db:"attempt_number"`
	NextRetryAt   *time.Time             `json:"next_retry_at,omitempty" db:"next_retry_at"`
	DeliveredAt   *time.Time             `json:"delivered_at,omitempty" db:"delivered_at"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// DeliveryStatus represents the status of a webhook delivery
type DeliveryStatus string

const (
	DeliveryStatusPending DeliveryStatus = "pending"
	DeliveryStatusSuccess  DeliveryStatus = "success"
	DeliveryStatusFailed   DeliveryStatus = "failed"
	DeliveryStatusRetrying DeliveryStatus = "retrying"
)

// WebhookPayload represents the structure of a webhook payload
type WebhookPayload struct {
	ID        string                 `json:"id"`
	EventType string                 `json:"event_type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

