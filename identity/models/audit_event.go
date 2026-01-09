package models

import (
	"time"

	"github.com/google/uuid"
)

// EventType constants for structured audit events
const (
	// User events
	EventTypeUserCreated   = "user.created"
	EventTypeUserUpdated   = "user.updated"
	EventTypeUserDeleted   = "user.deleted"
	EventTypeUserLocked    = "user.locked"
	EventTypeUserUnlocked  = "user.unlocked"
	EventTypeUserActivated = "user.activated"
	EventTypeUserDisabled  = "user.disabled"

	// Role events
	EventTypeRoleAssigned = "role.assigned"
	EventTypeRoleRemoved  = "role.removed"
	EventTypeRoleCreated  = "role.created"
	EventTypeRoleUpdated  = "role.updated"
	EventTypeRoleDeleted  = "role.deleted"

	// Permission events
	EventTypePermissionAssigned = "permission.assigned"
	EventTypePermissionRemoved  = "permission.removed"
	EventTypePermissionCreated  = "permission.created"
	EventTypePermissionUpdated  = "permission.updated"
	EventTypePermissionDeleted  = "permission.deleted"

	// MFA events
	EventTypeMFAEnrolled = "mfa.enrolled"
	EventTypeMFAVerified = "mfa.verified"
	EventTypeMFADisabled = "mfa.disabled"
	EventTypeMFAReset    = "mfa.reset"

	// Tenant events
	EventTypeTenantCreated   = "tenant.created"
	EventTypeTenantUpdated   = "tenant.updated"
	EventTypeTenantDeleted   = "tenant.deleted"
	EventTypeTenantSuspended  = "tenant.suspended"
	EventTypeTenantResumed   = "tenant.resumed"
	EventTypeTenantSettingsUpdated = "tenant.settings.updated"

	// Authentication events
	EventTypeLoginSuccess = "login.success"
	EventTypeLoginFailure = "login.failure"
	EventTypeTokenIssued  = "token.issued"
	EventTypeTokenRevoked = "token.revoked"
)

// Result constants
const (
	ResultSuccess = "success"
	ResultFailure = "failure"
	ResultDenied  = "denied"
)

// AuditEvent represents a structured audit event
type AuditEvent struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	EventType   string                 `json:"event_type" db:"event_type"`
	Actor       AuditActor             `json:"actor" db:"-"`
	Target      *AuditTarget           `json:"target,omitempty" db:"-"`
	Timestamp   time.Time              `json:"timestamp" db:"timestamp"`
	SourceIP    string                 `json:"source_ip,omitempty" db:"source_ip"`
	UserAgent   string                 `json:"user_agent,omitempty" db:"user_agent"`
	TenantID    *uuid.UUID             `json:"tenant_id,omitempty" db:"tenant_id"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	Result      string                 `json:"result" db:"result"`
	Error       string                 `json:"error,omitempty" db:"error"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`

	// Database fields (flattened for storage)
	ActorUserID        uuid.UUID `json:"-" db:"actor_user_id"`
	ActorUsername      string    `json:"-" db:"actor_username"`
	ActorPrincipalType string    `json:"-" db:"actor_principal_type"`
	TargetType         *string   `json:"-" db:"target_type"`
	TargetID           *uuid.UUID `json:"-" db:"target_id"`
	TargetIdentifier   *string   `json:"-" db:"target_identifier"`
}

// AuditActor represents who performed the action
type AuditActor struct {
	UserID        uuid.UUID `json:"user_id"`
	Username      string    `json:"username"`
	PrincipalType string    `json:"principal_type"` // "SYSTEM" or "TENANT"
}

// AuditTarget represents what was affected
type AuditTarget struct {
	Type       string    `json:"type"`       // "user", "role", "tenant", etc.
	ID         uuid.UUID `json:"id"`
	Identifier string    `json:"identifier"` // username, role name, etc.
}

// Flatten converts AuditEvent to database format
func (e *AuditEvent) Flatten() {
	e.ActorUserID = e.Actor.UserID
	e.ActorUsername = e.Actor.Username
	e.ActorPrincipalType = e.Actor.PrincipalType

	if e.Target != nil {
		e.TargetType = &e.Target.Type
		e.TargetID = &e.Target.ID
		if e.Target.Identifier != "" {
			e.TargetIdentifier = &e.Target.Identifier
		}
	}
}

// Expand converts database format to AuditEvent
func (e *AuditEvent) Expand() {
	e.Actor = AuditActor{
		UserID:        e.ActorUserID,
		Username:      e.ActorUsername,
		PrincipalType: e.ActorPrincipalType,
	}

	if e.TargetType != nil {
		e.Target = &AuditTarget{
			Type: *e.TargetType,
		}
		if e.TargetID != nil {
			e.Target.ID = *e.TargetID
		}
		if e.TargetIdentifier != nil {
			e.Target.Identifier = *e.TargetIdentifier
		}
	}
}

// Validate validates the audit event
func (e *AuditEvent) Validate() error {
	if e.EventType == "" {
		return ErrInvalidEventType
	}
	if e.Actor.UserID == uuid.Nil {
		return ErrInvalidActor
	}
	if e.Result != ResultSuccess && e.Result != ResultFailure && e.Result != ResultDenied {
		return ErrInvalidResult
	}
	return nil
}

// Errors
var (
	ErrInvalidEventType = &ValidationError{Message: "event type is required"}
	ErrInvalidActor     = &ValidationError{Message: "actor user ID is required"}
	ErrInvalidResult    = &ValidationError{Message: "result must be success, failure, or denied"}
)

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

