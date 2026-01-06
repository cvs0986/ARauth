package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	UserID      *uuid.UUID             `json:"user_id,omitempty"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	ResourceID  *string                `json:"resource_id,omitempty"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Status      string                 `json:"status"` // success, failure, error
	Message     string                 `json:"message,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// AuditRepository defines the interface for audit log storage
type AuditRepository interface {
	// Create creates a new audit log entry
	Create(ctx context.Context, log *AuditLog) error

	// List retrieves audit logs with filters
	List(ctx context.Context, tenantID uuid.UUID, filters *AuditFilters) ([]*AuditLog, error)

	// GetByID retrieves an audit log by ID
	GetByID(ctx context.Context, id uuid.UUID) (*AuditLog, error)
}

// AuditFilters represents filters for audit log queries
type AuditFilters struct {
	UserID     *uuid.UUID
	Action     *string
	Resource   *string
	Status     *string
	StartDate  *time.Time
	EndDate    *time.Time
	Page       int
	PageSize   int
}

