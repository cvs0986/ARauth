package interfaces

import (
	"context"

	"github.com/google/uuid"
)

// MFARecoveryCodeRepository defines the interface for MFA recovery code storage
type MFARecoveryCodeRepository interface {
	// CreateRecoveryCodes creates recovery codes for a user
	CreateRecoveryCodes(ctx context.Context, userID uuid.UUID, codes []string) error

	// GetRecoveryCodes retrieves recovery codes for a user
	GetRecoveryCodes(ctx context.Context, userID uuid.UUID) ([]string, error)

	// VerifyAndDeleteRecoveryCode verifies a recovery code and deletes it if valid
	VerifyAndDeleteRecoveryCode(ctx context.Context, userID uuid.UUID, code string) (bool, error)

	// DeleteRecoveryCodes deletes all recovery codes for a user
	DeleteRecoveryCodes(ctx context.Context, userID uuid.UUID) error
}

