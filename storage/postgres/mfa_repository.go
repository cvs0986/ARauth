package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/storage/interfaces"
)

// mfaRecoveryCodeRepository implements MFARecoveryCodeRepository for PostgreSQL
type mfaRecoveryCodeRepository struct {
	db *sql.DB
}

// NewMFARecoveryCodeRepository creates a new PostgreSQL MFA recovery code repository
func NewMFARecoveryCodeRepository(db *sql.DB) interfaces.MFARecoveryCodeRepository {
	return &mfaRecoveryCodeRepository{db: db}
}

// CreateRecoveryCodes creates recovery codes for a user
func (r *mfaRecoveryCodeRepository) CreateRecoveryCodes(ctx context.Context, userID uuid.UUID, codes []string) error {
	// Delete existing codes first
	if err := r.DeleteRecoveryCodes(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete existing codes: %w", err)
	}

	// Insert new codes
	query := `
		INSERT INTO mfa_recovery_codes (id, user_id, code_hash, used, created_at)
		VALUES (gen_random_uuid(), $1, encode(digest($2, 'sha256'), 'hex'), false, NOW())
	`

	for _, code := range codes {
		_, err := r.db.ExecContext(ctx, query, userID, code)
		if err != nil {
			return fmt.Errorf("failed to create recovery code: %w", err)
		}
	}

	return nil
}

// GetRecoveryCodes retrieves recovery codes for a user (returns hashed codes for verification only)
func (r *mfaRecoveryCodeRepository) GetRecoveryCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	// We don't return actual codes for security, only check if they exist
	query := `
		SELECT COUNT(*) FROM mfa_recovery_codes
		WHERE user_id = $1 AND used = false
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to get recovery codes: %w", err)
	}

	// Return empty slice - actual codes are never returned for security
	return []string{}, nil
}

// VerifyAndDeleteRecoveryCode verifies a recovery code and deletes it if valid
func (r *mfaRecoveryCodeRepository) VerifyAndDeleteRecoveryCode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	query := `
		UPDATE mfa_recovery_codes
		SET used = true, used_at = NOW()
		WHERE user_id = $1 
		  AND code_hash = encode(digest($2, 'sha256'), 'hex')
		  AND used = false
		RETURNING id
	`

	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, query, userID, code).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil // Code not found or already used
	}
	if err != nil {
		return false, fmt.Errorf("failed to verify recovery code: %w", err)
	}

	return true, nil
}

// DeleteRecoveryCodes deletes all recovery codes for a user
func (r *mfaRecoveryCodeRepository) DeleteRecoveryCodes(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM mfa_recovery_codes WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete recovery codes: %w", err)
	}

	return nil
}

// Helper function to hash recovery codes (not used directly, but for reference)
func hashRecoveryCode(code string) string {
	// This is done in SQL using digest() function
	// Keeping this for reference
	return ""
}

