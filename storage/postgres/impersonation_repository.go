package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/google/uuid"
)

// ImpersonationRepository implements ImpersonationRepository interface
type ImpersonationRepository struct {
	db *sql.DB
}

// NewImpersonationRepository creates a new impersonation repository
func NewImpersonationRepository(db *sql.DB) interfaces.ImpersonationRepository {
	return &ImpersonationRepository{db: db}
}

// Create creates a new impersonation session
func (r *ImpersonationRepository) Create(ctx context.Context, session *models.ImpersonationSession) error {
	query := `
		INSERT INTO impersonation_sessions (
			id, impersonator_user_id, target_user_id, tenant_id,
			started_at, token_jti, reason, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	metadataJSON, err := json.Marshal(session.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	now := time.Now()
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	if session.StartedAt.IsZero() {
		session.StartedAt = now
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	if session.UpdatedAt.IsZero() {
		session.UpdatedAt = now
	}

	var tokenJTIStr interface{}
	if session.TokenJTI != nil {
		tokenJTIStr = session.TokenJTI.String()
	}

	_, err = r.db.ExecContext(ctx, query,
		session.ID,
		session.ImpersonatorID,
		session.TargetUserID,
		session.TenantID,
		session.StartedAt,
		tokenJTIStr,
		session.Reason,
		metadataJSON,
		session.CreatedAt,
		session.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create impersonation session: %w", err)
	}

	return nil
}

// GetByID retrieves an impersonation session by ID
func (r *ImpersonationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ImpersonationSession, error) {
	query := `
		SELECT id, impersonator_user_id, target_user_id, tenant_id,
		       started_at, ended_at, token_jti, reason, metadata,
		       created_at, updated_at
		FROM impersonation_sessions
		WHERE id = $1
	`

	session := &models.ImpersonationSession{}
	var metadataJSON []byte
	var endedAt sql.NullTime
	var tokenJTIStr sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.ImpersonatorID,
		&session.TargetUserID,
		&session.TenantID,
		&session.StartedAt,
		&endedAt,
		&tokenJTIStr,
		&session.Reason,
		&metadataJSON,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("impersonation session not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get impersonation session: %w", err)
	}

	if endedAt.Valid {
		session.EndedAt = &endedAt.Time
	}

	if tokenJTIStr.Valid && tokenJTIStr.String != "" {
		if jti, err := uuid.Parse(tokenJTIStr.String); err == nil {
			session.TokenJTI = &jti
		}
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &session.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return session, nil
}

// GetByTokenJTI retrieves an impersonation session by token JTI
func (r *ImpersonationRepository) GetByTokenJTI(ctx context.Context, tokenJTI uuid.UUID) (*models.ImpersonationSession, error) {
	query := `
		SELECT id, impersonator_user_id, target_user_id, tenant_id,
		       started_at, ended_at, token_jti, reason, metadata,
		       created_at, updated_at
		FROM impersonation_sessions
		WHERE token_jti = $1 AND ended_at IS NULL
	`

	session := &models.ImpersonationSession{}
	var metadataJSON []byte
	var endedAt sql.NullTime
	var tokenJTIStr sql.NullString

	err := r.db.QueryRowContext(ctx, query, tokenJTI.String()).Scan(
		&session.ID,
		&session.ImpersonatorID,
		&session.TargetUserID,
		&session.TenantID,
		&session.StartedAt,
		&endedAt,
		&tokenJTIStr,
		&session.Reason,
		&metadataJSON,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("impersonation session not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get impersonation session: %w", err)
	}

	if endedAt.Valid {
		session.EndedAt = &endedAt.Time
	}

	if tokenJTIUUID.Valid {
		if jti, err := uuid.Parse(tokenJTIUUID.String); err == nil {
			session.TokenJTI = &jti
		}
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &session.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return session, nil
}

// GetActiveByImpersonator retrieves active impersonation sessions for an impersonator
func (r *ImpersonationRepository) GetActiveByImpersonator(ctx context.Context, impersonatorID uuid.UUID) ([]*models.ImpersonationSession, error) {
	query := `
		SELECT id, impersonator_user_id, target_user_id, tenant_id,
		       started_at, ended_at, token_jti, reason, metadata,
		       created_at, updated_at
		FROM impersonation_sessions
		WHERE impersonator_user_id = $1 AND ended_at IS NULL
		ORDER BY started_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, impersonatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to query impersonation sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*models.ImpersonationSession
	for rows.Next() {
		session := &models.ImpersonationSession{}
		var metadataJSON []byte
		var endedAt sql.NullTime
		var tokenJTIStr sql.NullString

		err := rows.Scan(
			&session.ID,
			&session.ImpersonatorID,
			&session.TargetUserID,
			&session.TenantID,
			&session.StartedAt,
			&endedAt,
			&tokenJTIStr,
			&session.Reason,
			&metadataJSON,
			&session.CreatedAt,
			&session.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan impersonation session: %w", err)
		}

		if endedAt.Valid {
			session.EndedAt = &endedAt.Time
		}

		if tokenJTIStr.Valid && tokenJTIStr.String != "" {
			if jti, err := uuid.Parse(tokenJTIStr.String); err == nil {
				session.TokenJTI = &jti
			}
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &session.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// GetActiveByTarget retrieves active impersonation sessions for a target user
func (r *ImpersonationRepository) GetActiveByTarget(ctx context.Context, targetUserID uuid.UUID) ([]*models.ImpersonationSession, error) {
	query := `
		SELECT id, impersonator_user_id, target_user_id, tenant_id,
		       started_at, ended_at, token_jti, reason, metadata,
		       created_at, updated_at
		FROM impersonation_sessions
		WHERE target_user_id = $1 AND ended_at IS NULL
		ORDER BY started_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to query impersonation sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*models.ImpersonationSession
	for rows.Next() {
		session := &models.ImpersonationSession{}
		var metadataJSON []byte
		var endedAt sql.NullTime
		var tokenJTIStr sql.NullString

		err := rows.Scan(
			&session.ID,
			&session.ImpersonatorID,
			&session.TargetUserID,
			&session.TenantID,
			&session.StartedAt,
			&endedAt,
			&tokenJTIStr,
			&session.Reason,
			&metadataJSON,
			&session.CreatedAt,
			&session.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan impersonation session: %w", err)
		}

		if endedAt.Valid {
			session.EndedAt = &endedAt.Time
		}

		if tokenJTIStr.Valid && tokenJTIStr.String != "" {
			if jti, err := uuid.Parse(tokenJTIStr.String); err == nil {
				session.TokenJTI = &jti
			}
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &session.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// EndSession ends an impersonation session
func (r *ImpersonationRepository) EndSession(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE impersonation_sessions
		SET ended_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND ended_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to end impersonation session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("impersonation session not found or already ended")
	}

	return nil
}

// List lists impersonation sessions with filters
func (r *ImpersonationRepository) List(ctx context.Context, filters *interfaces.ImpersonationFilters) ([]*models.ImpersonationSession, error) {
	query := `
		SELECT id, impersonator_user_id, target_user_id, tenant_id,
		       started_at, ended_at, token_jti, reason, metadata,
		       created_at, updated_at
		FROM impersonation_sessions
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	if filters != nil {
		if filters.ImpersonatorID != nil {
			query += fmt.Sprintf(" AND impersonator_user_id = $%d", argPos)
			args = append(args, *filters.ImpersonatorID)
			argPos++
		}

		if filters.TargetUserID != nil {
			query += fmt.Sprintf(" AND target_user_id = $%d", argPos)
			args = append(args, *filters.TargetUserID)
			argPos++
		}

		if filters.TenantID != nil {
			query += fmt.Sprintf(" AND tenant_id = $%d", argPos)
			args = append(args, *filters.TenantID)
			argPos++
		}

		if filters.ActiveOnly {
			query += " AND ended_at IS NULL"
		}
	}

	query += " ORDER BY started_at DESC"

	if filters != nil && filters.PageSize > 0 {
		offset := (filters.Page - 1) * filters.PageSize
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
		args = append(args, filters.PageSize, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query impersonation sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*models.ImpersonationSession
	for rows.Next() {
		session := &models.ImpersonationSession{}
		var metadataJSON []byte
		var endedAt sql.NullTime
		var tokenJTIStr sql.NullString

		err := rows.Scan(
			&session.ID,
			&session.ImpersonatorID,
			&session.TargetUserID,
			&session.TenantID,
			&session.StartedAt,
			&endedAt,
			&tokenJTIStr,
			&session.Reason,
			&metadataJSON,
			&session.CreatedAt,
			&session.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan impersonation session: %w", err)
		}

		if endedAt.Valid {
			session.EndedAt = &endedAt.Time
		}

		if tokenJTIStr.Valid && tokenJTIStr.String != "" {
			if jti, err := uuid.Parse(tokenJTIStr.String); err == nil {
				session.TokenJTI = &jti
			}
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &session.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

