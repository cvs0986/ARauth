package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/credential"
	"github.com/nuage-identity/iam/storage/interfaces"
)

// credentialRepository implements CredentialRepository for PostgreSQL
type credentialRepository struct {
	db *sql.DB
}

// NewCredentialRepository creates a new PostgreSQL credential repository
func NewCredentialRepository(db *sql.DB) interfaces.CredentialRepository {
	return &credentialRepository{db: db}
}

// Create creates a new credential
func (r *credentialRepository) Create(ctx context.Context, cred *credential.Credential) error {
	query := `
		INSERT INTO credentials (
			id, user_id, password_hash, password_changed_at,
			password_expires_at, failed_login_attempts, locked_until,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now()
	if cred.ID == uuid.Nil {
		cred.ID = uuid.New()
	}
	if cred.CreatedAt.IsZero() {
		cred.CreatedAt = now
	}
	if cred.UpdatedAt.IsZero() {
		cred.UpdatedAt = now
	}
	if cred.PasswordChangedAt.IsZero() {
		cred.PasswordChangedAt = now
	}

	_, err := r.db.ExecContext(ctx, query,
		cred.ID, cred.UserID, cred.PasswordHash, cred.PasswordChangedAt,
		cred.PasswordExpiresAt, cred.FailedLoginAttempts, cred.LockedUntil,
		cred.CreatedAt, cred.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}

	return nil
}

// GetByUserID retrieves credentials by user ID
func (r *credentialRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*credential.Credential, error) {
	query := `
		SELECT id, user_id, password_hash, password_changed_at,
		       password_expires_at, failed_login_attempts, locked_until,
		       created_at, updated_at
		FROM credentials
		WHERE user_id = $1
	`

	cred := &credential.Credential{}
	var passwordExpiresAt, lockedUntil sql.NullTime

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&cred.ID, &cred.UserID, &cred.PasswordHash, &cred.PasswordChangedAt,
		&passwordExpiresAt, &cred.FailedLoginAttempts, &lockedUntil,
		&cred.CreatedAt, &cred.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("credential not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	if passwordExpiresAt.Valid {
		cred.PasswordExpiresAt = &passwordExpiresAt.Time
	}
	if lockedUntil.Valid {
		cred.LockedUntil = &lockedUntil.Time
	}

	return cred, nil
}

// Update updates existing credentials
func (r *credentialRepository) Update(ctx context.Context, cred *credential.Credential) error {
	query := `
		UPDATE credentials
		SET password_hash = $2, password_changed_at = $3,
		    password_expires_at = $4, failed_login_attempts = $5,
		    locked_until = $6, updated_at = $7
		WHERE user_id = $1
	`

	cred.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		cred.UserID, cred.PasswordHash, cred.PasswordChangedAt,
		cred.PasswordExpiresAt, cred.FailedLoginAttempts, cred.LockedUntil,
		cred.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update credential: %w", err)
	}

	return nil
}

// Delete deletes credentials
func (r *credentialRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM credentials WHERE user_id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("credential not found")
	}

	return nil
}

