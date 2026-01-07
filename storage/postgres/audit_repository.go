package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// auditRepository implements AuditRepository for PostgreSQL
type auditRepository struct {
	db *sql.DB
}

// NewAuditRepository creates a new PostgreSQL audit repository
func NewAuditRepository(db *sql.DB) interfaces.AuditRepository {
	return &auditRepository{db: db}
}

// Create creates a new audit log entry
func (r *auditRepository) Create(ctx context.Context, log *interfaces.AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			id, tenant_id, user_id, action, resource, resource_id,
			ip_address, user_agent, status, message, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	var metadataJSON []byte
	if log.Metadata != nil {
		metadataJSON, _ = json.Marshal(log.Metadata)
	}

	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.TenantID, log.UserID, log.Action, log.Resource,
		log.ResourceID, log.IPAddress, log.UserAgent, log.Status,
		log.Message, metadataJSON, log.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetByID retrieves an audit log by ID
func (r *auditRepository) GetByID(ctx context.Context, id uuid.UUID) (*interfaces.AuditLog, error) {
	query := `
		SELECT id, tenant_id, user_id, action, resource, resource_id,
		       ip_address, user_agent, status, message, metadata, created_at
		FROM audit_logs
		WHERE id = $1
	`

	log := &interfaces.AuditLog{}
	var userID sql.NullString
	var resourceID sql.NullString
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID, &log.TenantID, &userID, &log.Action, &log.Resource,
		&resourceID, &log.IPAddress, &log.UserAgent, &log.Status,
		&log.Message, &metadataJSON, &log.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("audit log not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	if userID.Valid {
		uid, _ := uuid.Parse(userID.String)
		log.UserID = &uid
	}
	if resourceID.Valid {
		log.ResourceID = &resourceID.String
	}
	if len(metadataJSON) > 0 {
		_ = json.Unmarshal(metadataJSON, &log.Metadata) // Ignore unmarshal errors for optional metadata
	}

	return log, nil
}

// List retrieves audit logs with filters
func (r *auditRepository) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.AuditFilters) ([]*interfaces.AuditLog, error) {
	if filters == nil {
		filters = &interfaces.AuditFilters{
			Page:     1,
			PageSize: 50,
		}
	}

	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 || filters.PageSize > 100 {
		filters.PageSize = 50
	}

	offset := (filters.Page - 1) * filters.PageSize

	query := `
		SELECT id, tenant_id, user_id, action, resource, resource_id,
		       ip_address, user_agent, status, message, metadata, created_at
		FROM audit_logs
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argPos := 2

	if filters.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, *filters.UserID)
		argPos++
	}

	if filters.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", argPos)
		args = append(args, *filters.Action)
		argPos++
	}

	if filters.Resource != nil {
		query += fmt.Sprintf(" AND resource = $%d", argPos)
		args = append(args, *filters.Resource)
		argPos++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *filters.Status)
		argPos++
	}

	if filters.StartDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argPos)
		args = append(args, *filters.StartDate)
		argPos++
	}

	if filters.EndDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argPos)
		args = append(args, *filters.EndDate)
		argPos++
	}

	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", argPos) + " OFFSET $" + fmt.Sprintf("%d", argPos+1)
	args = append(args, filters.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*interfaces.AuditLog
	for rows.Next() {
		log := &interfaces.AuditLog{}
		var userID, resourceID sql.NullString
		var metadataJSON []byte

		err := rows.Scan(
			&log.ID, &log.TenantID, &userID, &log.Action, &log.Resource,
			&resourceID, &log.IPAddress, &log.UserAgent, &log.Status,
			&log.Message, &metadataJSON, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if userID.Valid {
			uid, _ := uuid.Parse(userID.String)
			log.UserID = &uid
		}
		if resourceID.Valid {
			log.ResourceID = &resourceID.String
		}
		if len(metadataJSON) > 0 {
			_ = json.Unmarshal(metadataJSON, &log.Metadata) // Ignore unmarshal errors for optional metadata
		}

		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating audit logs: %w", err)
	}

	return logs, nil
}

