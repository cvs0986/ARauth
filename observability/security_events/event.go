package security_events

import (
	"time"

	"github.com/google/uuid"
)

// EventType defines the type of security event
type EventType string

const (
	EventAuthFailure           EventType = "auth_failure"
	EventPermissionDenied      EventType = "permission_denied"
	EventRateLimitExceeded     EventType = "rate_limit_exceeded"
	EventTokenValidationFailed EventType = "token_validation_failed"
	EventBlacklistedTokenUsed  EventType = "blacklisted_token_used"
	EventFederationFailure     EventType = "federation_failure"
	EventMFAFailure            EventType = "mfa_failure"
	EventSuspiciousActivity    EventType = "suspicious_activity"
	EventUserCreated           EventType = "user_created"
	EventUserDeleted           EventType = "user_deleted"
	EventRoleAssigned          EventType = "role_assigned"
	EventRoleRevoked           EventType = "role_revoked"
)

// Severity defines the severity level of a security event
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// SecurityEvent represents a security-related event in the system
type SecurityEvent struct {
	ID        uuid.UUID              `json:"id"`
	EventType EventType              `json:"event_type"`
	Severity  Severity               `json:"severity"`
	TenantID  *uuid.UUID             `json:"tenant_id,omitempty"`
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
	IP        string                 `json:"ip,omitempty"`
	Resource  string                 `json:"resource,omitempty"`
	Action    string                 `json:"action,omitempty"`
	Result    string                 `json:"result,omitempty"` // success, failure, denied
	Details   map[string]interface{} `json:"details,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// EventFilters defines filters for querying security events
type EventFilters struct {
	EventType *EventType
	Severity  *Severity
	TenantID  *uuid.UUID
	UserID    *uuid.UUID
	IP        *string
	Since     *time.Time
	Until     *time.Time
	Limit     int
	Offset    int
}

// NewSecurityEvent creates a new security event with default values
func NewSecurityEvent(eventType EventType, severity Severity) *SecurityEvent {
	return &SecurityEvent{
		ID:        uuid.New(),
		EventType: eventType,
		Severity:  severity,
		Details:   make(map[string]interface{}),
		CreatedAt: time.Now(),
	}
}

// WithTenant sets the tenant ID for the event
func (e *SecurityEvent) WithTenant(tenantID uuid.UUID) *SecurityEvent {
	e.TenantID = &tenantID
	return e
}

// WithUser sets the user ID for the event
func (e *SecurityEvent) WithUser(userID uuid.UUID) *SecurityEvent {
	e.UserID = &userID
	return e
}

// WithIP sets the IP address for the event
func (e *SecurityEvent) WithIP(ip string) *SecurityEvent {
	e.IP = ip
	return e
}

// WithResource sets the resource for the event
func (e *SecurityEvent) WithResource(resource string) *SecurityEvent {
	e.Resource = resource
	return e
}

// WithAction sets the action for the event
func (e *SecurityEvent) WithAction(action string) *SecurityEvent {
	e.Action = action
	return e
}

// WithResult sets the result for the event
func (e *SecurityEvent) WithResult(result string) *SecurityEvent {
	e.Result = result
	return e
}

// WithDetail adds a detail to the event
func (e *SecurityEvent) WithDetail(key string, value interface{}) *SecurityEvent {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithDetails sets multiple details for the event
func (e *SecurityEvent) WithDetails(details map[string]interface{}) *SecurityEvent {
	e.Details = details
	return e
}
