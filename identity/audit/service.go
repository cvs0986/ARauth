package audit

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/identity/webhook"
	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/google/uuid"
)

// Service provides audit event logging functionality
type Service struct {
	repo           interfaces.AuditEventRepository
	webhookService webhook.ServiceInterface
}

// NewService creates a new audit service
func NewService(repo interfaces.AuditEventRepository, webhookService webhook.ServiceInterface) ServiceInterface {
	return &Service{
		repo:           repo,
		webhookService: webhookService,
	}
}

// LogEvent logs a structured audit event
func (s *Service) LogEvent(ctx context.Context, event *models.AuditEvent) error {
	if err := event.Validate(); err != nil {
		return fmt.Errorf("invalid audit event: %w", err)
	}

	// Save audit event
	if err := s.repo.Create(ctx, event); err != nil {
		return err
	}

	// Trigger webhooks asynchronously (don't block on webhook delivery)
	go s.triggerWebhooks(context.Background(), event)

	return nil
}

// triggerWebhooks triggers webhooks for an audit event
func (s *Service) triggerWebhooks(ctx context.Context, event *models.AuditEvent) {
	if s.webhookService == nil {
		return
	}

	// Convert audit event to webhook payload
	payload := s.eventToPayload(event)

	// Get tenant ID (may be nil for system events)
	var tenantID uuid.UUID
	if event.TenantID != nil {
		tenantID = *event.TenantID
	} else {
		// System events don't trigger tenant webhooks
		return
	}

	// Trigger webhook (async, errors are logged by webhook service)
	_ = s.webhookService.TriggerWebhook(ctx, tenantID, event.EventType, payload, &event.ID)
}

// eventToPayload converts an audit event to a webhook payload map
func (s *Service) eventToPayload(event *models.AuditEvent) map[string]interface{} {
	// Marshal event to JSON and unmarshal to map to get all fields
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return map[string]interface{}{
			"event_type": event.EventType,
			"timestamp":  event.Timestamp,
			"error":      "failed to serialize event",
		}
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(eventJSON, &payload); err != nil {
		return map[string]interface{}{
			"event_type": event.EventType,
			"timestamp":  event.Timestamp,
			"error":      "failed to deserialize event",
		}
	}

	return payload
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

// ExportEvents exports audit events as CSV based on filters
func (s *Service) ExportEvents(ctx context.Context, filters *interfaces.AuditEventFilters) ([]byte, string, error) {
	// Query events (reuse existing query logic)
	// For export, we might want to increase the limit, but for now we respect the filters
	// If filters.PageSize is 0, QueryEvents might use default.
	// Ideally for export we want "all matching" or a large limit.
	// Let's assume the caller sets appropriate pagination or we iterate.
	// For a V1 MVP, we will fetch up to MaxPageSize (100) or whatever is requested.
	// To support full export, we would need pagination loop, but let's start simple as per requirement.

	events, _, err := s.repo.QueryEvents(ctx, filters)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query audit events for export: %w", err)
	}

	b := &bytes.Buffer{}
	w := csv.NewWriter(b)

	// Write Header
	header := []string{"Event ID", "Timestamp", "Event Type", "Actor", "Result", "IP Address", "Target Type", "Target ID"}
	if err := w.Write(header); err != nil {
		return nil, "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write Rows
	for _, event := range events {
		event.Expand() // Populate Actor/Target details if needed

		record := []string{
			event.ID.String(),
			event.Timestamp.Format(time.RFC3339),
			event.EventType,
			fmt.Sprintf("%s (%s)", event.Actor.Username, event.Actor.PrincipalType),
			event.Result,
			event.SourceIP,
			"", "",
		}

		if event.Target != nil {
			record[6] = event.Target.Type
			if event.Target.ID != uuid.Nil {
				record[7] = event.Target.ID.String()
			}
		}

		if err := w.Write(record); err != nil {
			return nil, "", fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, "", fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	filename := fmt.Sprintf("audit_export_%s.csv", time.Now().Format("20060102_150405"))
	return b.Bytes(), filename, nil
}

// createEvent is a helper to create an audit event
func (s *Service) createEvent(eventType string, actor models.AuditActor, target *models.AuditTarget, tenantID *uuid.UUID, sourceIP, userAgent string, result string, errorMsg string, metadata map[string]interface{}) *models.AuditEvent {
	event := &models.AuditEvent{
		ID:        uuid.New(),
		EventType: eventType,
		Actor:     actor,
		Target:    target,
		Timestamp: time.Now(),
		SourceIP:  sourceIP,
		UserAgent: userAgent,
		TenantID:  tenantID,
		Metadata:  metadata,
		Result:    result,
		Error:     errorMsg,
		CreatedAt: time.Now(),
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

// LogMFAChallengeCreated logs an MFA challenge creation event
func (s *Service) LogMFAChallengeCreated(ctx context.Context, actor models.AuditActor, tenantID *uuid.UUID, sourceIP, userAgent string) error {
	event := s.createEvent(models.EventTypeMFAChallengeCreated, actor, nil, tenantID, sourceIP, userAgent, models.ResultSuccess, "", nil)
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
