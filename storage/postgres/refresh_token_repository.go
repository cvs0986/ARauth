package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/google/uuid"
)

// refreshTokenRepository implements RefreshTokenRepository for PostgreSQL
type refreshTokenRepository struct {
	db *sql.DB
}

// NewRefreshTokenRepository creates a new PostgreSQL refresh token repository
func NewRefreshTokenRepository(db *sql.DB) interfaces.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

// Create creates a new refresh token
func (r *refreshTokenRepository) Create(ctx context.Context, token *interfaces.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (
			id, user_id, tenant_id, token_hash, expires_at, revoked_at,
			remember_me, mfa_verified, created_at, updated_at
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

	// Handle nullable tenant_id for SYSTEM users
	var tenantIDValue interface{}
	if token.TenantID != uuid.Nil {
		tenantIDValue = token.TenantID
	} else {
		tenantIDValue = nil
	}

	_, err := r.db.ExecContext(ctx, query,
		token.ID, token.UserID, tenantIDValue, token.TokenHash,
		token.ExpiresAt, token.RevokedAt, token.RememberMe, token.MFAVerified,
		token.CreatedAt, token.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

// GetByTokenHash retrieves a refresh token by its hash
func (r *refreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*interfaces.RefreshToken, error) {
	query := `
		SELECT id, user_id, tenant_id, token_hash, expires_at, revoked_at,
		       remember_me, mfa_verified, created_at, updated_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	token := &interfaces.RefreshToken{}
	var revokedAt sql.NullTime
	var tenantID sql.NullString

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID, &token.UserID, &tenantID, &token.TokenHash,
		&token.ExpiresAt, &revokedAt, &token.RememberMe, &token.MFAVerified,
		&token.CreatedAt, &token.UpdatedAt,
	)

	// Handle nullable tenant_id
	if tenantID.Valid {
		parsedTenantID, err := uuid.Parse(tenantID.String)
		if err == nil {
			token.TenantID = parsedTenantID
		}
	}

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("refresh token not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	if revokedAt.Valid {
		token.RevokedAt = &revokedAt.Time
	}

	return token, nil
}

// GetByUserID retrieves all active refresh tokens for a user
func (r *refreshTokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*interfaces.RefreshToken, error) {
	query := `
		SELECT id, user_id, tenant_id, token_hash, expires_at, revoked_at,
		       remember_me, mfa_verified, created_at, updated_at
		FROM refresh_tokens
		WHERE user_id = $1 AND (revoked_at IS NULL OR revoked_at > NOW())
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*interfaces.RefreshToken
	for rows.Next() {
		token := &interfaces.RefreshToken{}
		var revokedAt sql.NullTime
		var tenantID sql.NullString

		err := rows.Scan(
			&token.ID, &token.UserID, &tenantID, &token.TokenHash,
			&token.ExpiresAt, &revokedAt, &token.RememberMe, &token.MFAVerified,
			&token.CreatedAt, &token.UpdatedAt,
		)

		// Handle nullable tenant_id
		if tenantID.Valid {
			parsedTenantID, err := uuid.Parse(tenantID.String)
			if err == nil {
				token.TenantID = parsedTenantID
			}
		}
		if err != nil {
			return nil, fmt.Errorf("failed to scan refresh token: %w", err)
		}

		if revokedAt.Valid {
			token.RevokedAt = &revokedAt.Time
		}

		tokens = append(tokens, token)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating refresh tokens: %w", err)
	}

	return tokens, nil
}

// Revoke revokes a refresh token
func (r *refreshTokenRepository) Revoke(ctx context.Context, tokenID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND revoked_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("refresh token not found or already revoked")
	}

	return nil
}

// RevokeByTokenHash revokes a refresh token by its hash
func (r *refreshTokenRepository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW(), updated_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("refresh token not found or already revoked")
	}

	return nil
}

// RevokeAllForUser revokes all refresh tokens for a user
func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all refresh tokens: %w", err)
	}

	return nil
}

// DeleteExpired deletes expired refresh tokens
func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < NOW() AND (revoked_at IS NULL OR revoked_at < NOW())
	`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	return nil
}
