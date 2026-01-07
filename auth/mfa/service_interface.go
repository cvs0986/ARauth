package mfa

import (
	"context"

	"github.com/google/uuid"
)

// ServiceInterface defines the interface for MFA service operations
type ServiceInterface interface {
	Enroll(ctx context.Context, req *EnrollRequest) (*EnrollResponse, error)
	Verify(ctx context.Context, req *VerifyRequest) (bool, error)
	CreateChallenge(ctx context.Context, req *ChallengeRequest) (*ChallengeResponse, error)
	VerifyChallenge(ctx context.Context, req *VerifyChallengeRequest) (*VerifyChallengeResponse, error)
}

