package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// SCIMTokenRepository implements SCIMTokenRepository interface
type SCIMTokenRepository struct {
	db *sql.DB
}

// NewSCIMTokenRepository creates a new SCIM token repository
func NewSCIMTokenRepository(db *sql.DB) interfaces.SCIMTokenRepository {
	return &SCIMTokenRepository{db: db}
}

// Create creates a new SCIM token
func (r *SCIMTokenRepository) Create(ctx context.Context, token *models.SCIMToken) error {
	query := `
		INSERT INTO scim_tokens (
			id, tenant_id, name, token_hash, lookup_hash, scopes, expires_at,
			created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	if token.CreatedAt.IsZero() {
		token.CreatedAt = now
	}
	if token.UpdatedAt.IsZero() {
		token.UpdatedAt = now
	}

	_, err := r.db.ExecContext(ctx, query,
		token.ID,
		token.TenantID,
		token.Name,
		token.TokenHash,
		token.LookupHash,
		pq.Array(token.Scopes),
		token.ExpiresAt,
		token.CreatedBy,
		token.CreatedAt,
		token.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create SCIM token: %w", err)
	}

	return nil
}

// GetByID retrieves a SCIM token by ID
func (r *SCIMTokenRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SCIMToken, error) {
	query := `
		SELECT id, tenant_id, name, token_hash, lookup_hash, scopes, expires_at,
		       last_used_at, created_by, created_at, updated_at, deleted_at
		FROM scim_tokens
		WHERE id = $1 AND deleted_at IS NULL
	`

	token := &models.SCIMToken{}
	var scopes pq.StringArray
	var expiresAt, lastUsedAt, deletedAt sql.NullTime
	var createdBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&token.ID,
		&token.TenantID,
		&token.Name,
		&token.TokenHash,
		&token.LookupHash,
		&scopes,
		&expiresAt,
		&lastUsedAt,
		&createdBy,
		&token.CreatedAt,
		&token.UpdatedAt,
		&deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("SCIM token not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get SCIM token: %w", err)
	}

	token.Scopes = []string(scopes)
	if expiresAt.Valid {
		token.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		token.LastUsedAt = &lastUsedAt.Time
	}
	if deletedAt.Valid {
		token.DeletedAt = &deletedAt.Time
	}
	if createdBy.Valid {
		if createdByUUID, err := uuid.Parse(createdBy.String); err == nil {
			token.CreatedBy = &createdByUUID
		}
	}

	return token, nil
}

// GetByLookupHash retrieves a SCIM token by its lookup hash (SHA256)
func (r *SCIMTokenRepository) GetByLookupHash(ctx context.Context, lookupHash string) (*models.SCIMToken, error) {
	query := `
		SELECT id, tenant_id, name, token_hash, lookup_hash, scopes, expires_at,
		       last_used_at, created_by, created_at, updated_at, deleted_at
		FROM scim_tokens
		WHERE lookup_hash = $1 AND deleted_at IS NULL
	`

	token := &models.SCIMToken{}
	var scopes pq.StringArray
	var expiresAt, lastUsedAt, deletedAt sql.NullTime
	var createdBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, lookupHash).Scan(
		&token.ID,
		&token.TenantID,
		&token.Name,
		&token.TokenHash,
		&token.LookupHash,
		&scopes,
		&expiresAt,
		&lastUsedAt,
		&createdBy,
		&token.CreatedAt,
		&token.UpdatedAt,
		&deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("SCIM token not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get SCIM token: %w", err)
	}

	token.Scopes = []string(scopes)
	if expiresAt.Valid {
		token.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		token.LastUsedAt = &lastUsedAt.Time
	}
	if deletedAt.Valid {
		token.DeletedAt = &deletedAt.Time
	}
	if createdBy.Valid {
		if createdByUUID, err := uuid.Parse(createdBy.String); err == nil {
			token.CreatedBy = &createdByUUID
		}
	}

	return token, nil
}

// List lists SCIM tokens for a tenant
func (r *SCIMTokenRepository) List(ctx context.Context, tenantID uuid.UUID) ([]*models.SCIMToken, error) {
	query := `
		SELECT id, tenant_id, name, token_hash, lookup_hash, scopes, expires_at,
		       last_used_at, created_by, created_at, updated_at, deleted_at
		FROM scim_tokens
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query SCIM tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*models.SCIMToken
	for rows.Next() {
		token := &models.SCIMToken{}
		var scopes pq.StringArray
		var expiresAt, lastUsedAt, deletedAt sql.NullTime
		var createdBy sql.NullString

		err := rows.Scan(
			&token.ID,
			&token.TenantID,
			&token.Name,
			&token.TokenHash,
			&token.LookupHash,
			&scopes,
			&expiresAt,
			&lastUsedAt,
			&createdBy,
			&token.CreatedAt,
			&token.UpdatedAt,
			&deletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan SCIM token: %w", err)
		}

		token.Scopes = []string(scopes)
		if expiresAt.Valid {
			token.ExpiresAt = &expiresAt.Time
		}
		if lastUsedAt.Valid {
			token.LastUsedAt = &lastUsedAt.Time
		}
		if deletedAt.Valid {
			token.DeletedAt = &deletedAt.Time
		}
		if createdBy.Valid {
			if createdByUUID, err := uuid.Parse(createdBy.String); err == nil {
				token.CreatedBy = &createdByUUID
			}
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

// Update updates a SCIM token
func (r *SCIMTokenRepository) Update(ctx context.Context, token *models.SCIMToken) error {
	query := `
		UPDATE scim_tokens
		SET name = $1, scopes = $2, expires_at = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`

	token.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		token.Name,
		pq.Array(token.Scopes),
		token.ExpiresAt,
		token.UpdatedAt,
		token.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update SCIM token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("SCIM token not found or already deleted")
	}

	return nil
}

// Delete soft-deletes a SCIM token
func (r *SCIMTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE scim_tokens
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete SCIM token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("SCIM token not found or already deleted")
	}

	return nil
}

// UpdateLastUsed updates the last_used_at timestamp
func (r *SCIMTokenRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE scim_tokens
		SET last_used_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update last used timestamp: %w", err)
	}

	return nil
}

