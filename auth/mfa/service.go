package mfa

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/security/encryption"
	"github.com/arauth-identity/iam/security/totp"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// Service provides MFA functionality
type Service struct {
	userRepo            interfaces.UserRepository
	credentialRepo      interfaces.CredentialRepository
	mfaRecoveryCodeRepo interfaces.MFARecoveryCodeRepository
	totpGenerator       *totp.Generator
	encryptor           *encryption.Encryptor
	sessionManager      *SessionManager
}

// NewService creates a new MFA service
func NewService(
	userRepo interfaces.UserRepository,
	credentialRepo interfaces.CredentialRepository,
	mfaRecoveryCodeRepo interfaces.MFARecoveryCodeRepository,
	totpGenerator *totp.Generator,
	encryptor *encryption.Encryptor,
	sessionManager *SessionManager,
) *Service {
	return &Service{
		userRepo:            userRepo,
		credentialRepo:      credentialRepo,
		mfaRecoveryCodeRepo: mfaRecoveryCodeRepo,
		totpGenerator:       totpGenerator,
		encryptor:           encryptor,
		sessionManager:      sessionManager,
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

	// Encrypt and store TOTP secret
	encryptedSecret, err := s.encryptor.Encrypt(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt TOTP secret: %w", err)
	}

	// Update user with encrypted secret and enable MFA
	user.MFAEnabled = true
	user.MFASecretEncrypted = &encryptedSecret
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user with MFA secret: %w", err)
	}

	// Store recovery codes (hashed in database)
	if err := s.mfaRecoveryCodeRepo.CreateRecoveryCodes(ctx, user.ID, recoveryCodes); err != nil {
		return nil, fmt.Errorf("failed to store recovery codes: %w", err)
	}

	return &EnrollResponse{
		Secret:        secret, // Return plaintext secret only once for QR code setup
		QRCode:        qrCodeBase64,
		RecoveryCodes: recoveryCodes, // Return recovery codes only once
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

	if req.TOTPCode != "" {
		// Get and decrypt TOTP secret
		if user.MFASecretEncrypted == nil {
			return false, fmt.Errorf("MFA secret not found")
		}

		secret, err := s.encryptor.Decrypt(*user.MFASecretEncrypted)
		if err != nil {
			return false, fmt.Errorf("failed to decrypt TOTP secret: %w", err)
		}

		// Verify TOTP code
		valid := s.totpGenerator.Validate(secret, req.TOTPCode)
		return valid, nil
	}

	if req.RecoveryCode != "" {
		// Verify recovery code
		valid, err := s.mfaRecoveryCodeRepo.VerifyAndDeleteRecoveryCode(ctx, req.UserID, req.RecoveryCode)
		if err != nil {
			return false, fmt.Errorf("failed to verify recovery code: %w", err)
		}
		return valid, nil
	}

	return false, fmt.Errorf("either totp_code or recovery_code must be provided")
}

