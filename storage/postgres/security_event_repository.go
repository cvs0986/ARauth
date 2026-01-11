package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/arauth-identity/iam/observability/security_events"
)

// SecurityEventRepository implements the security_events.Repository interface
type SecurityEventRepository struct {
	db *sql.DB
}

// NewSecurityEventRepository creates a new security event repository
func NewSecurityEventRepository(db *sql.DB) *SecurityEventRepository {
	return &SecurityEventRepository{db: db}
}

// Create stores a new security event
func (r *SecurityEventRepository) Create(ctx context.Context, event *security_events.SecurityEvent) error {
	detailsJSON, err := json.Marshal(event.Details)
	if err != nil {
		return fmt.Errorf("failed to marshal details: %w", err)
	}

	query := `
		INSERT INTO security_events (
			id, event_type, severity, tenant_id, user_id, ip,
			resource, action, result, details, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.ExecContext(ctx, query,
		event.ID,
		event.EventType,
		event.Severity,
		event.TenantID,
		event.UserID,
		event.IP,
		event.Resource,
		event.Action,
		event.Result,
		detailsJSON,
		event.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create security event: %w", err)
	}

	return nil
}

// CreateBatch stores multiple security events in a single transaction
func (r *SecurityEventRepository) CreateBatch(ctx context.Context, events []*security_events.SecurityEvent) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO security_events (
			id, event_type, severity, tenant_id, user_id, ip,
			resource, action, result, details, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, event := range events {
		detailsJSON, err := json.Marshal(event.Details)
		if err != nil {
			return fmt.Errorf("failed to marshal details: %w", err)
		}

		_, err = stmt.ExecContext(ctx,
			event.ID,
			event.EventType,
			event.Severity,
			event.TenantID,
			event.UserID,
			event.IP,
			event.Resource,
			event.Action,
			event.Result,
			detailsJSON,
			event.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Find retrieves security events with filters
func (r *SecurityEventRepository) Find(ctx context.Context, filters security_events.EventFilters) ([]*security_events.SecurityEvent, error) {
	query := `
		SELECT id, event_type, severity, tenant_id, user_id, ip,
		       resource, action, result, details, created_at
		FROM security_events
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	// Build WHERE clause
	if filters.EventType != nil {
		query += fmt.Sprintf(" AND event_type = $%d", argCount)
		args = append(args, *filters.EventType)
		argCount++
	}
	if filters.Severity != nil {
		query += fmt.Sprintf(" AND severity = $%d", argCount)
		args = append(args, *filters.Severity)
		argCount++
	}
	if filters.TenantID != nil {
		query += fmt.Sprintf(" AND tenant_id = $%d", argCount)
		args = append(args, *filters.TenantID)
		argCount++
	}
	if filters.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *filters.UserID)
		argCount++
	}
	if filters.IP != nil {
		query += fmt.Sprintf(" AND ip = $%d", argCount)
		args = append(args, *filters.IP)
		argCount++
	}
	if filters.Since != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, *filters.Since)
		argCount++
	}
	if filters.Until != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, *filters.Until)
		argCount++
	}

	// Order by created_at DESC
	query += " ORDER BY created_at DESC"

	// Add limit and offset
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.Limit)
		argCount++
	}
	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filters.Offset)
		argCount++
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query security events: %w", err)
	}
	defer rows.Close()

	events := []*security_events.SecurityEvent{}
	for rows.Next() {
		event := &security_events.SecurityEvent{}
		var detailsJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.Severity,
			&event.TenantID,
			&event.UserID,
			&event.IP,
			&event.Resource,
			&event.Action,
			&event.Result,
			&detailsJSON,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if len(detailsJSON) > 0 {
			if err := json.Unmarshal(detailsJSON, &event.Details); err != nil {
				return nil, fmt.Errorf("failed to unmarshal details: %w", err)
			}
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return events, nil
}

// Count returns the count of events matching filters
func (r *SecurityEventRepository) Count(ctx context.Context, filters security_events.EventFilters) (int, error) {
	query := "SELECT COUNT(*) FROM security_events WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	// Build WHERE clause (same as Find)
	if filters.EventType != nil {
		query += fmt.Sprintf(" AND event_type = $%d", argCount)
		args = append(args, *filters.EventType)
		argCount++
	}
	if filters.Severity != nil {
		query += fmt.Sprintf(" AND severity = $%d", argCount)
		args = append(args, *filters.Severity)
		argCount++
	}
	if filters.TenantID != nil {
		query += fmt.Sprintf(" AND tenant_id = $%d", argCount)
		args = append(args, *filters.TenantID)
		argCount++
	}
	if filters.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *filters.UserID)
		argCount++
	}
	if filters.IP != nil {
		query += fmt.Sprintf(" AND ip = $%d", argCount)
		args = append(args, *filters.IP)
		argCount++
	}
	if filters.Since != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, *filters.Since)
		argCount++
	}
	if filters.Until != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, *filters.Until)
		argCount++
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count security events: %w", err)
	}

	return count, nil
}

// DeleteOlderThan deletes events older than the specified time
func (r *SecurityEventRepository) DeleteOlderThan(ctx context.Context, olderThan time.Time) (int, error) {
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM security_events WHERE created_at < $1",
		olderThan,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old security events: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}
