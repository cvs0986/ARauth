package mfa

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// ChallengeRequest represents a request to create an MFA challenge
type ChallengeRequest struct {
	UserID   uuid.UUID `json:"user_id" binding:"required"`
	TenantID uuid.UUID `json:"tenant_id" binding:"required"`
}

// ChallengeResponse represents the response from creating an MFA challenge
type ChallengeResponse struct {
	SessionID string `json:"session_id"`
	ExpiresIn int    `json:"expires_in"` // seconds
}

// CreateChallenge creates an MFA challenge session
// If user hasn't enrolled yet but MFA is required, this will allow enrollment flow
func (s *Service) CreateChallenge(ctx context.Context, req *ChallengeRequest) (*ChallengeResponse, error) {
	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Verify tenant matches (skip for SYSTEM users where tenantID is uuid.Nil)
	if req.TenantID != uuid.Nil {
		if user.TenantID == nil || *user.TenantID != req.TenantID {
			return nil, fmt.Errorf("tenant mismatch")
		}
	} else {
		// For SYSTEM users, verify they don't have a tenant_id
		if user.TenantID != nil {
			return nil, fmt.Errorf("tenant mismatch: user has tenant but SYSTEM user expected")
		}
	}

	// If user hasn't enrolled in MFA yet, allow challenge creation for enrollment flow
	// The enrollment will happen via the enroll endpoint, and then verification will enable MFA
	// This allows the login flow to proceed with enrollment when tenant requires MFA

	// Create MFA session
	sessionID, err := s.sessionManager.CreateSession(ctx, req.UserID, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create MFA session: %w", err)
	}

	return &ChallengeResponse{
		SessionID: sessionID,
		ExpiresIn: 300, // 5 minutes in seconds
	}, nil
}

// VerifyChallenge verifies an MFA challenge with TOTP or recovery code
type VerifyChallengeRequest struct {
	SessionID   string `json:"session_id" binding:"required"`
	TOTPCode    string `json:"totp_code,omitempty"`
	RecoveryCode string `json:"recovery_code,omitempty"`
}

// VerifyChallengeResponse represents the response from verifying an MFA challenge
type VerifyChallengeResponse struct {
	Verified bool   `json:"verified"`
	UserID   string `json:"user_id,omitempty"`
	TenantID string `json:"tenant_id,omitempty"`
}

// VerifyChallenge verifies an MFA challenge
func (s *Service) VerifyChallenge(ctx context.Context, req *VerifyChallengeRequest) (*VerifyChallengeResponse, error) {
	// Get and verify session
	session, err := s.sessionManager.VerifySession(ctx, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired session: %w", err)
	}

	// Increment attempts
	if err := s.sessionManager.IncrementAttempts(ctx, req.SessionID); err != nil {
		return nil, err
	}

	// Verify MFA code
	verifyReq := &VerifyRequest{
		UserID:      session.UserID,
		TOTPCode:    req.TOTPCode,
		RecoveryCode: req.RecoveryCode,
	}

	valid, err := s.Verify(ctx, verifyReq)
	if err != nil {
		return nil, err
	}

	if !valid {
		return &VerifyChallengeResponse{
			Verified: false,
		}, nil
	}

	// Delete session on successful verification
	_ = s.sessionManager.DeleteSession(ctx, req.SessionID) // Ignore error on cleanup

	// Build response - handle uuid.Nil for SYSTEM users
	response := &VerifyChallengeResponse{
		Verified: true,
		UserID:   session.UserID.String(),
	}
	if session.TenantID != uuid.Nil {
		tenantIDStr := session.TenantID.String()
		response.TenantID = tenantIDStr
	}

	return response, nil
}

