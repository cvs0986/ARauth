package mfa

import (
	"context"

	"github.com/google/uuid"
)

// ServiceInterface defines the interface for MFA service operations
type ServiceInterface interface {
	Enroll(ctx context.Context, req *EnrollRequest) (*EnrollResponse, error)
	Verify(ctx context.Context, req *VerifyRequest) (bool, error)
	GenerateRecoveryCodes(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]string, error)
	VerifyRecoveryCode(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, code string) error
	CreateChallenge(ctx context.Context, req *ChallengeRequest) (*ChallengeResponse, error)
	VerifyChallenge(ctx context.Context, req *VerifyChallengeRequest) (*VerifyChallengeResponse, error)
}

