package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// auditEventRepository implements AuditEventRepository for PostgreSQL
type auditEventRepository struct {
	db *sql.DB
}

// NewAuditEventRepository creates a new PostgreSQL audit event repository
func NewAuditEventRepository(db *sql.DB) interfaces.AuditEventRepository {
	return &auditEventRepository{db: db}
}

// Create creates a new audit event
func (r *auditEventRepository) Create(ctx context.Context, event *models.AuditEvent) error {
	// Flatten the event for database storage
	event.Flatten()

	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	var metadataJSON []byte
	if event.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(event.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		INSERT INTO audit_events (
			id, event_type, actor_user_id, actor_username, actor_principal_type,
			target_type, target_id, target_identifier, timestamp, source_ip,
			user_agent, tenant_id, metadata, result, error, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err := r.db.ExecContext(ctx, query,
		event.ID,
		event.EventType,
		event.ActorUserID,
		event.ActorUsername,
		event.ActorPrincipalType,
		event.TargetType,
		event.TargetID,
		event.TargetIdentifier,
		event.Timestamp,
		event.SourceIP,
		event.UserAgent,
		event.TenantID,
		metadataJSON,
		event.Result,
		event.Error,
		event.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit event: %w", err)
	}

	// Expand back for consistency
	event.Expand()

	return nil
}

// QueryEvents retrieves audit events with filters
func (r *auditEventRepository) QueryEvents(ctx context.Context, filters *interfaces.AuditEventFilters) ([]*models.AuditEvent, int, error) {
	filters.Validate()

	var conditions []string
	var args []interface{}
	argIndex := 1

	// Build WHERE clause
	if filters.EventType != nil {
		conditions = append(conditions, fmt.Sprintf("event_type = $%d", argIndex))
		args = append(args, *filters.EventType)
		argIndex++
	}

	if filters.ActorUserID != nil {
		conditions = append(conditions, fmt.Sprintf("actor_user_id = $%d", argIndex))
		args = append(args, *filters.ActorUserID)
		argIndex++
	}

	if filters.TargetType != nil {
		conditions = append(conditions, fmt.Sprintf("target_type = $%d", argIndex))
		args = append(args, *filters.TargetType)
		argIndex++
	}

	if filters.TargetID != nil {
		conditions = append(conditions, fmt.Sprintf("target_id = $%d", argIndex))
		args = append(args, *filters.TargetID)
		argIndex++
	}

	if filters.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
		args = append(args, *filters.TenantID)
		argIndex++
	}

	if filters.Result != nil {
		conditions = append(conditions, fmt.Sprintf("result = $%d", argIndex))
		args = append(args, *filters.Result)
		argIndex++
	}

	if filters.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argIndex))
		args = append(args, *filters.StartDate)
		argIndex++
	}

	if filters.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", argIndex))
		args = append(args, *filters.EndDate)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_events %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit events: %w", err)
	}

	// Query with pagination
	offset := (filters.Page - 1) * filters.PageSize
	query := fmt.Sprintf(`
		SELECT 
			id, event_type, actor_user_id, actor_username, actor_principal_type,
			target_type, target_id, target_identifier, timestamp, source_ip,
			user_agent, tenant_id, metadata, result, error, created_at
		FROM audit_events
		%s
		ORDER BY timestamp DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, filters.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit events: %w", err)
	}
	defer rows.Close()

	var events []*models.AuditEvent
	for rows.Next() {
		event := &models.AuditEvent{}
		var metadataJSON []byte
		var sourceIP sql.NullString
		var userAgent sql.NullString
		var errorMsg sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.ActorUserID,
			&event.ActorUsername,
			&event.ActorPrincipalType,
			&event.TargetType,
			&event.TargetID,
			&event.TargetIdentifier,
			&event.Timestamp,
			&sourceIP,
			&userAgent,
			&event.TenantID,
			&metadataJSON,
			&event.Result,
			&errorMsg,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit event: %w", err)
		}

		// Expand the event
		event.Expand()

		// Set nullable fields
		if sourceIP.Valid {
			event.SourceIP = sourceIP.String
		}
		if userAgent.Valid {
			event.UserAgent = userAgent.String
		}
		if errorMsg.Valid {
			event.Error = errorMsg.String
		}

		// Parse metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating audit events: %w", err)
	}

	return events, total, nil
}

// GetEvent retrieves an audit event by ID
func (r *auditEventRepository) GetEvent(ctx context.Context, eventID uuid.UUID) (*models.AuditEvent, error) {
	query := `
		SELECT 
			id, event_type, actor_user_id, actor_username, actor_principal_type,
			target_type, target_id, target_identifier, timestamp, source_ip,
			user_agent, tenant_id, metadata, result, error, created_at
		FROM audit_events
		WHERE id = $1
	`

	event := &models.AuditEvent{}
	var metadataJSON []byte
	var sourceIP sql.NullString
	var userAgent sql.NullString
	var errorMsg sql.NullString

	err := r.db.QueryRowContext(ctx, query, eventID).Scan(
		&event.ID,
		&event.EventType,
		&event.ActorUserID,
		&event.ActorUsername,
		&event.ActorPrincipalType,
		&event.TargetType,
		&event.TargetID,
		&event.TargetIdentifier,
		&event.Timestamp,
		&sourceIP,
		&userAgent,
		&event.TenantID,
		&metadataJSON,
		&event.Result,
		&errorMsg,
		&event.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("audit event not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get audit event: %w", err)
	}

	// Expand the event
	event.Expand()

	// Set nullable fields
	if sourceIP.Valid {
		event.SourceIP = sourceIP.String
	}
	if userAgent.Valid {
		event.UserAgent = userAgent.String
	}
	if errorMsg.Valid {
		event.Error = errorMsg.String
	}

	// Parse metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return event, nil
}

