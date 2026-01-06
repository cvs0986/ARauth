package mfa

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/security/totp"
	"github.com/nuage-identity/iam/storage/interfaces"
)

// Service provides MFA functionality
type Service struct {
	userRepo       interfaces.UserRepository
	credentialRepo interfaces.CredentialRepository
	totpGenerator  *totp.Generator
}

// NewService creates a new MFA service
func NewService(
	userRepo interfaces.UserRepository,
	credentialRepo interfaces.CredentialRepository,
	totpGenerator *totp.Generator,
) *Service {
	return &Service{
		userRepo:       userRepo,
		credentialRepo: credentialRepo,
		totpGenerator:  totpGenerator,
	}
}

// EnrollRequest represents a request to enroll in MFA
type EnrollRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

// EnrollResponse represents the response from MFA enrollment
type EnrollResponse struct {
	Secret    string   `json:"secret"`
	QRCode    string   `json:"qr_code"` // Base64 encoded PNG
	RecoveryCodes []string `json:"recovery_codes"`
}

// VerifyRequest represents a request to verify MFA
type VerifyRequest struct {
	UserID      uuid.UUID `json:"user_id"`
	TOTPCode    string    `json:"totp_code,omitempty"`
	RecoveryCode string   `json:"recovery_code,omitempty"`
}

// Enroll enrolls a user in MFA
func (s *Service) Enroll(ctx context.Context, req *EnrollRequest) (*EnrollResponse, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Generate TOTP secret
	secret, err := s.totpGenerator.GenerateSecret(user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	// Generate QR code
	qrCodeBytes, err := s.totpGenerator.GenerateQRCode(user.Email, secret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Encode QR code as base64
	qrCodeBase64 := fmt.Sprintf("data:image/png;base64,%s", 
		base64.StdEncoding.EncodeToString(qrCodeBytes))

	// Generate recovery codes
	recoveryCodes, err := s.totpGenerator.GenerateRecoveryCodes(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recovery codes: %w", err)
	}

	// TODO: Store secret and recovery codes (encrypted) in database
	// For now, return them to the client

	return &EnrollResponse{
		Secret:        secret,
		QRCode:        qrCodeBase64,
		RecoveryCodes: recoveryCodes,
	}, nil
}

// Verify verifies a TOTP code or recovery code
func (s *Service) Verify(ctx context.Context, req *VerifyRequest) (bool, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return false, fmt.Errorf("user not found: %w", err)
	}

	// Check if MFA is enabled
	if !user.MFAEnabled {
		return false, fmt.Errorf("MFA is not enabled for this user")
	}

	// TODO: Get encrypted secret from database
	// For now, this is a placeholder

	if req.TOTPCode != "" {
		// Verify TOTP code
		// secret := getSecretFromDB(user.ID)
		// return s.totpGenerator.Validate(secret, req.TOTPCode), nil
		return false, fmt.Errorf("TOTP verification not yet implemented with database storage")
	}

	if req.RecoveryCode != "" {
		// Verify recovery code
		// return verifyRecoveryCode(user.ID, req.RecoveryCode), nil
		return false, fmt.Errorf("recovery code verification not yet implemented with database storage")
	}

	return false, fmt.Errorf("either totp_code or recovery_code must be provided")
}

