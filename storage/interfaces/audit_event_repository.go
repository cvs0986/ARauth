package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
)

// AuditEventRepository defines the interface for structured audit event storage
type AuditEventRepository interface {
	// Create creates a new audit event
	Create(ctx context.Context, event *models.AuditEvent) error

	// QueryEvents retrieves audit events with filters
	QueryEvents(ctx context.Context, filters *AuditEventFilters) ([]*models.AuditEvent, int, error)

	// GetEvent retrieves an audit event by ID
	GetEvent(ctx context.Context, eventID uuid.UUID) (*models.AuditEvent, error)
}

// AuditEventFilters represents filters for audit event queries
type AuditEventFilters struct {
	EventType   *string
	ActorUserID *uuid.UUID
	TargetType  *string
	TargetID    *uuid.UUID
	TenantID    *uuid.UUID
	Result      *string
	StartDate   *time.Time
	EndDate     *time.Time
	Page        int
	PageSize    int
}

// DefaultPageSize is the default page size for pagination
const DefaultPageSize = 50

// MaxPageSize is the maximum page size allowed
const MaxPageSize = 1000

// Validate validates and normalizes the filters
func (f *AuditEventFilters) Validate() {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = DefaultPageSize
	}
	if f.PageSize > MaxPageSize {
		f.PageSize = MaxPageSize
	}
}

