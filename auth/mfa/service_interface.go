package mfa

import (
	"context"

	"github.com/google/uuid"
)

// ServiceInterface defines the interface for MFA service operations
type ServiceInterface interface {
	Enroll(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) (*EnrollResponse, error)
	Verify(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, code string, sessionID string) error
	Disable(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) error
	GenerateRecoveryCodes(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]string, error)
	VerifyRecoveryCode(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, code string) error
	Challenge(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) (*ChallengeResponse, error)
}

