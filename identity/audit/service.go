package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// Service provides audit event logging functionality
type Service struct {
	repo interfaces.AuditEventRepository
}

// NewService creates a new audit service
func NewService(repo interfaces.AuditEventRepository) ServiceInterface {
	return &Service{
		repo: repo,
	}
}

// LogEvent logs a structured audit event
func (s *Service) LogEvent(ctx context.Context, event *models.AuditEvent) error {
	if err := event.Validate(); err != nil {
		return fmt.Errorf("invalid audit event: %w", err)
	}

	return s.repo.Create(ctx, event)
}

// QueryEvents queries audit events with filters
func (s *Service) QueryEvents(ctx context.Context, filters *interfaces.AuditEventFilters) ([]*models.AuditEvent, int, error) {
	events, total, err := s.repo.QueryEvents(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit events: %w", err)
	}

	// Expand all events
	for _, event := range events {
		event.Expand()
	}

	return events, total, nil
}

// GetEvent retrieves a specific audit event by ID
func (s *Service) GetEvent(ctx context.Context, eventID uuid.UUID) (*models.AuditEvent, error) {
	event, err := s.repo.GetEvent(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit event: %w", err)
	}

	event.Expand()
	return event, nil
}

// createEvent is a helper to create an audit event
func (s *Service) createEvent(eventType string, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string, result string, errorMsg string, metadata map[string]interface{}) *models.AuditEvent {
	event := &models.AuditEvent{
		ID:          uuid.New(),
		EventType:   eventType,
		Actor:       actor,
		Target:      target,
		Timestamp:   time.Now(),
		SourceIP:    sourceIP,
		UserAgent:   userAgent,
		TenantID:    tenantID,
		Metadata:    metadata,
		Result:      result,
		Error:       errorMsg,
		CreatedAt:   time.Now(),
	}
	return event
}

// LogUserCreated logs a user creation event
func (s *Service) LogUserCreated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error {
	event := s.createEvent(models.EventTypeUserCreated, actor, target, tenantID, sourceIP, userAgent, models.ResultSuccess, "", metadata)
	return s.LogEvent(ctx, event)
}

// LogUserUpdated logs a user update event
func (s *Service) LogUserUpdated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error {
	event := s.createEvent(models.EventTypeUserUpdated, actor, target, tenantID, sourceIP, userAgent, models.ResultSuccess, "", metadata)
	return s.LogEvent(ctx, event)
}

// LogUserDeleted logs a user deletion event
func (s *Service) LogUserDeleted(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypeUserDeleted, actor, target, tenantID, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogUserLocked logs a user lock event
func (s *Service) LogUserLocked(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypeUserLocked, actor, target, tenantID, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogUserUnlocked logs a user unlock event
func (s *Service) LogUserUnlocked(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypeUserUnlocked, actor, target, tenantID, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogRoleAssigned logs a role assignment event
func (s *Service) LogRoleAssigned(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error {
	event := s.createEvent(models.EventTypeRoleAssigned, actor, target, tenantID, sourceIP, userAgent, models.ResultSuccess, "", metadata)
	return s.LogEvent(ctx, event)
}

// LogRoleRemoved logs a role removal event
func (s *Service) LogRoleRemoved(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypeRoleRemoved, actor, target, tenantID, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogRoleCreated logs a role creation event
func (s *Service) LogRoleCreated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypeRoleCreated, actor, target, tenantID, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogRoleUpdated logs a role update event
func (s *Service) LogRoleUpdated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypeRoleUpdated, actor, target, tenantID, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogRoleDeleted logs a role deletion event
func (s *Service) LogRoleDeleted(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypeRoleDeleted, actor, target, tenantID, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogPermissionAssigned logs a permission assignment event
func (s *Service) LogPermissionAssigned(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypePermissionAssigned, actor, target, tenantID, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogPermissionRemoved logs a permission removal event
func (s *Service) LogPermissionRemoved(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypePermissionRemoved, actor, target, tenantID, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogPermissionCreated logs a permission creation event
func (s *Service) LogPermissionCreated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypePermissionCreated, actor, target, tenantID, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogPermissionUpdated logs a permission update event
func (s *Service) LogPermissionUpdated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypePermissionUpdated, actor, target, tenantID, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogPermissionDeleted logs a permission deletion event
func (s *Service) LogPermissionDeleted(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypePermissionDeleted, actor, target, tenantID, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogMFAEnrolled logs an MFA enrollment event
func (s *Service) LogMFAEnrolled(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypeMFAEnrolled, actor, nil, tenantID, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogMFAVerified logs an MFA verification event
func (s *Service) LogMFAVerified(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string, success bool) error {
	result := models.ResultSuccess
	if !success {
		result = models.ResultFailure
	}
	event := s.createEvent(models.EventTypeMFAVerified, actor, nil, tenantID, sourceIP, userAgent, result, "", nil)
	return s.LogEvent(ctx, event)
}

// LogMFADisabled logs an MFA disable event
func (s *Service) LogMFADisabled(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypeMFADisabled, actor, nil, tenantID, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogMFAReset logs an MFA reset event
func (s *Service) LogMFAReset(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypeMFAReset, actor, target, tenantID, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogTenantCreated logs a tenant creation event
func (s *Service) LogTenantCreated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypeTenantCreated, actor, target, nil, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogTenantUpdated logs a tenant update event
func (s *Service) LogTenantUpdated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypeTenantUpdated, actor, target, nil, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogTenantDeleted logs a tenant deletion event
func (s *Service) LogTenantDeleted(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypeTenantDeleted, actor, target, nil, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogTenantSuspended logs a tenant suspension event
func (s *Service) LogTenantSuspended(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypeTenantSuspended, actor, target, nil, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogTenantResumed logs a tenant resumption event
func (s *Service) LogTenantResumed(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypeTenantResumed, actor, target, nil, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogTenantSettingsUpdated logs a tenant settings update event
func (s *Service) LogTenantSettingsUpdated(ctx context.Context, actor models.AuditActor, target *models.AuditTarget, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypeTenantSettingsUpdated, actor, target, nil, sourceIP, userAgent, models.ResultSuccess, "", nil)
	return s.LogEvent(ctx, event)
}

// LogLoginSuccess logs a successful login event
func (s *Service) LogLoginSuccess(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error {
	event := s.createEvent(models.EventTypeLoginSuccess, actor, nil, tenantID, sourceIP, userAgent, models.ResultSuccess, "", metadata)
	return s.LogEvent(ctx, event)
}

// LogLoginFailure logs a failed login event
func (s *Service) LogLoginFailure(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string, reason string) error {
	event := s.createEvent(models.EventTypeLoginFailure, actor, nil, tenantID, sourceIP, userAgent, models.ResultFailure, reason, nil)
	return s.LogEvent(ctx, event)
}

// LogTokenIssued logs a token issuance event
func (s *Service) LogTokenIssued(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error {
	event := s.createEvent(models.EventTypeTokenIssued, actor, nil, tenantID, sourceIP, userAgent, models.ResultSuccess, "", metadata)
	return s.LogEvent(ctx, event)
}

// LogTokenRevoked logs a token revocation event
func (s *Service) LogTokenRevoked(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string, metadata map[string]interface{}) error {
	event := s.createEvent(models.EventTypeTokenRevoked, actor, nil, tenantID, sourceIP, userAgent, models.ResultSuccess, "", metadata)
	return s.LogEvent(ctx, event)
}

