package mfa

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/arauth-identity/iam/identity/capability"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/security/encryption"
	"github.com/arauth-identity/iam/security/totp"
	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/google/uuid"
)

// Service provides MFA functionality
type Service struct {
	userRepo            interfaces.UserRepository
	credentialRepo      interfaces.CredentialRepository
	mfaRecoveryCodeRepo interfaces.MFARecoveryCodeRepository
	totpGenerator       *totp.Generator
	encryptor           *encryption.Encryptor
	sessionManager      *SessionManager
	capabilityService   capability.ServiceInterface
}

// NewService creates a new MFA service
func NewService(
	userRepo interfaces.UserRepository,
	credentialRepo interfaces.CredentialRepository,
	mfaRecoveryCodeRepo interfaces.MFARecoveryCodeRepository,
	totpGenerator *totp.Generator,
	encryptor *encryption.Encryptor,
	sessionManager *SessionManager,
	capabilityService capability.ServiceInterface,
) *Service {
	return &Service{
		userRepo:            userRepo,
		credentialRepo:      credentialRepo,
		mfaRecoveryCodeRepo: mfaRecoveryCodeRepo,
		totpGenerator:       totpGenerator,
		encryptor:           encryptor,
		sessionManager:      sessionManager,
		capabilityService:   capabilityService,
	}
}

// EnrollRequest represents a request to enroll in MFA
type EnrollRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

// EnrollResponse represents the response from MFA enrollment
type EnrollResponse struct {
	Secret        string   `json:"secret"`
	QRCode        string   `json:"qr_code"` // Base64 encoded PNG
	RecoveryCodes []string `json:"recovery_codes"`
}

// VerifyRequest represents a request to verify MFA
type VerifyRequest struct {
	UserID       uuid.UUID `json:"user_id"`
	TOTPCode     string    `json:"totp_code,omitempty"`
	RecoveryCode string    `json:"recovery_code,omitempty"`
}

// Enroll enrolls a user in MFA
func (s *Service) Enroll(ctx context.Context, req *EnrollRequest) (*EnrollResponse, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if MFA/TOTP is allowed and enabled via capability model
	var tenantID uuid.UUID
	if user.TenantID != nil {
		tenantID = *user.TenantID
		eval, err := s.capabilityService.EvaluateCapability(ctx, tenantID, user.ID, models.CapabilityKeyTOTP)
		if err != nil {
			return nil, fmt.Errorf("failed to check TOTP capability: %w", err)
		}
		if !eval.CanUse {
			return nil, fmt.Errorf("TOTP is not available for this tenant: %s", eval.Reason)
		}
	} else {
		// For SYSTEM users, check if TOTP is supported
		supported, err := s.capabilityService.IsCapabilitySupported(ctx, models.CapabilityKeyTOTP)
		if err != nil {
			return nil, fmt.Errorf("failed to check TOTP capability: %w", err)
		}
		if !supported {
			return nil, fmt.Errorf("TOTP is not supported")
		}
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

	// Store encrypted secret but DO NOT enable MFA yet
	// MFA will be enabled only after successful verification
	user.MFASecretEncrypted = &encryptedSecret
	// Keep MFAEnabled as false - it will be set to true after verification
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

// EnrollForLogin enrolls a user in MFA during login using a challenge session for security
func (s *Service) EnrollForLogin(ctx context.Context, sessionID string) (*EnrollResponse, error) {
	// Verify session exists and is valid
	session, err := s.sessionManager.VerifySession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired session: %w", err)
	}

	// Use the user ID from the session
	req := &EnrollRequest{
		UserID: session.UserID,
	}

	// Call the regular Enroll method
	return s.Enroll(ctx, req)
}

// Verify verifies a TOTP code or recovery code
// This method also enables MFA after first successful verification
func (s *Service) Verify(ctx context.Context, req *VerifyRequest) (bool, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return false, fmt.Errorf("user not found: %w", err)
	}

	// Check if MFA/TOTP is allowed and enabled via capability model
	var tenantID uuid.UUID
	if user.TenantID != nil {
		tenantID = *user.TenantID
		eval, err := s.capabilityService.EvaluateCapability(ctx, tenantID, user.ID, models.CapabilityKeyTOTP)
		if err != nil {
			return false, fmt.Errorf("failed to check TOTP capability: %w", err)
		}
		if !eval.CanUse {
			return false, fmt.Errorf("TOTP is not available for this tenant: %s", eval.Reason)
		}
	} else {
		// For SYSTEM users, check if TOTP is supported
		supported, err := s.capabilityService.IsCapabilitySupported(ctx, models.CapabilityKeyTOTP)
		if err != nil {
			return false, fmt.Errorf("failed to check TOTP capability: %w", err)
		}
		if !supported {
			return false, fmt.Errorf("TOTP is not supported")
		}
	}

	// Check if MFA secret exists (MFA may be enrolled but not yet verified)
	if user.MFASecretEncrypted == nil {
		return false, fmt.Errorf("MFA secret not found. Please enroll in MFA first.")
	}

	if req.TOTPCode != "" {
		// Get and decrypt TOTP secret
		secret, err := s.encryptor.Decrypt(*user.MFASecretEncrypted)
		if err != nil {
			return false, fmt.Errorf("failed to decrypt TOTP secret: %w", err)
		}

		// Verify TOTP code
		valid := s.totpGenerator.Validate(secret, req.TOTPCode)

		// If verification is successful and MFA is not yet enabled, enable it now
		// This allows enrollment without immediate verification, but requires verification before MFA is active
		if valid && !user.MFAEnabled {
			user.MFAEnabled = true
			if err := s.userRepo.Update(ctx, user); err != nil {
				// Log error but don't fail verification - MFA is already working
				// The flag will be set on next successful verification
				fmt.Printf("Warning: Failed to enable MFA flag after verification: %v\n", err)
			}
		}

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

// CreateSession creates a new MFA session
func (s *Service) CreateSession(ctx context.Context, userID, tenantID uuid.UUID) (string, error) {
	return s.sessionManager.CreateSession(ctx, userID, tenantID)
}
