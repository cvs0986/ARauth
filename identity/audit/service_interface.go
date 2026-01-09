package audit

import (
	"context"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// ServiceInterface defines the interface for audit service
type ServiceInterface interface {
	// LogEvent logs a structured audit event
	LogEvent(ctx context.Context, event *AuditEvent) error

	// QueryEvents queries audit events with filters
	QueryEvents(ctx context.Context, filters *interfaces.AuditEventFilters) ([]*AuditEvent, int, error)

	// GetEvent retrieves a specific audit event by ID
	GetEvent(ctx context.Context, eventID uuid.UUID) (*AuditEvent, error)

	// Helper methods for common events
	LogUserCreated(ctx context.Context, actor AuditActor, target *AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error
	LogUserUpdated(ctx context.Context, actor AuditActor, target *AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error
	LogUserDeleted(ctx context.Context, actor AuditActor, target *AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error
	LogUserLocked(ctx context.Context, actor AuditActor, target *AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error
	LogUserUnlocked(ctx context.Context, actor AuditActor, target *AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error

	LogRoleAssigned(ctx context.Context, actor AuditActor, target *AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error
	LogRoleRemoved(ctx context.Context, actor AuditActor, target *AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error
	LogRoleCreated(ctx context.Context, actor AuditActor, target *AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error
	LogRoleUpdated(ctx context.Context, actor AuditActor, target *AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error
	LogRoleDeleted(ctx context.Context, actor AuditActor, target *AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error

	LogPermissionAssigned(ctx context.Context, actor AuditActor, target *AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error
	LogPermissionRemoved(ctx context.Context, actor AuditActor, target *AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error
	LogPermissionCreated(ctx context.Context, actor AuditActor, target *AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error
	LogPermissionUpdated(ctx context.Context, actor AuditActor, target *AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error
	LogPermissionDeleted(ctx context.Context, actor AuditActor, target *AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error

	LogMFAEnrolled(ctx context.Context, actor AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string) error
	LogMFAVerified(ctx context.Context, actor AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string, success bool) error
	LogMFADisabled(ctx context.Context, actor AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string) error
	LogMFAReset(ctx context.Context, actor AuditActor, target *AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error

	LogTenantCreated(ctx context.Context, actor AuditActor, target *AuditTarget, sourceIP, userAgent string) error
	LogTenantUpdated(ctx context.Context, actor AuditActor, target *AuditTarget, sourceIP, userAgent string) error
	LogTenantDeleted(ctx context.Context, actor AuditActor, target *AuditTarget, sourceIP, userAgent string) error
	LogTenantSuspended(ctx context.Context, actor AuditActor, target *AuditTarget, sourceIP, userAgent string) error
	LogTenantResumed(ctx context.Context, actor AuditActor, target *AuditTarget, sourceIP, userAgent string) error
	LogTenantSettingsUpdated(ctx context.Context, actor AuditActor, target *AuditTarget, sourceIP, userAgent string) error

	LogLoginSuccess(ctx context.Context, actor AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error
	LogLoginFailure(ctx context.Context, actor AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string, reason string) error
	LogTokenIssued(ctx context.Context, actor AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error
	LogTokenRevoked(ctx context.Context, actor AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error
}

