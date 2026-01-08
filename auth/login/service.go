package login

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/auth/hydra"
	"github.com/arauth-identity/iam/auth/token"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/security/password"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// Service provides login functionality
type Service struct {
	userRepo            interfaces.UserRepository
	credentialRepo      interfaces.CredentialRepository
	refreshTokenRepo    interfaces.RefreshTokenRepository
	tenantSettingsRepo  interfaces.TenantSettingsRepository
	hydraClient         *hydra.Client
	passwordHasher      *password.Hasher
	claimsBuilder       *claims.Builder
	tokenService        token.ServiceInterface
	lifetimeResolver    *token.LifetimeResolver
}

// NewService creates a new login service
func NewService(
	userRepo interfaces.UserRepository,
	credentialRepo interfaces.CredentialRepository,
	refreshTokenRepo interfaces.RefreshTokenRepository,
	tenantSettingsRepo interfaces.TenantSettingsRepository,
	hydraClient *hydra.Client,
	claimsBuilder *claims.Builder,
	tokenService token.ServiceInterface,
	lifetimeResolver *token.LifetimeResolver,
) *Service {
	return &Service{
		userRepo:           userRepo,
		credentialRepo:     credentialRepo,
		refreshTokenRepo:   refreshTokenRepo,
		tenantSettingsRepo: tenantSettingsRepo,
		hydraClient:        hydraClient,
		passwordHasher:     password.NewHasher(),
		claimsBuilder:      claimsBuilder,
		tokenService:       tokenService,
		lifetimeResolver:   lifetimeResolver,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username       string    `json:"username" binding:"required"`
	Password       string    `json:"password" binding:"required"`
	TenantID       uuid.UUID `json:"tenant_id"` // Set from context, not from request body
	RememberMe    bool      `json:"remember_me,omitempty"` // Remember Me option
	LoginChallenge *string   `json:"login_challenge,omitempty"` // For OAuth2 flow
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken      string `json:"access_token,omitempty"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	IDToken          string `json:"id_token,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	ExpiresIn        int    `json:"expires_in,omitempty"`         // Access token expiry in seconds
	RefreshExpiresIn int    `json:"refresh_expires_in,omitempty"` // Refresh token expiry in seconds
	RememberMe       bool   `json:"remember_me,omitempty"`
	MFARequired      bool   `json:"mfa_required"`
	MFASessionID     string `json:"mfa_session_id,omitempty"`
	RedirectTo       string `json:"redirect_to,omitempty"` // For OAuth2 flow
}

// Login authenticates a user and returns tokens
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	var user *models.User
	var err error

	// Priority: If tenant ID is provided, try TENANT user first
	// If no tenant ID, try SYSTEM user
	if req.TenantID != uuid.Nil {
		// Tenant ID provided - try to find TENANT user first
		user, err = s.userRepo.GetByUsername(ctx, req.Username, req.TenantID)
		if err != nil || user == nil {
			// User not found in tenant - return generic error for security
			// Don't reveal whether username or tenant is wrong
			return nil, fmt.Errorf("invalid credentials")
		}
		
		// Verify the user is actually a TENANT user
		if user.PrincipalType != models.PrincipalTypeTenant {
			// User exists but is not a TENANT user (might be SYSTEM user)
			// Return generic error for security
			return nil, fmt.Errorf("invalid credentials")
		}
		
		// Verify tenant ID matches
		if user.TenantID == nil || *user.TenantID != req.TenantID {
			return nil, fmt.Errorf("invalid credentials")
		}
	} else {
		// No tenant ID provided - try SYSTEM user
		// First, try to get as SYSTEM user by username
		systemUser, systemErr := s.userRepo.GetSystemUserByUsername(ctx, req.Username)
		if systemErr == nil && systemUser != nil && systemUser.PrincipalType == models.PrincipalTypeSystem {
			user = systemUser
		} else {
			// If username lookup failed, try by email (username might be email)
			systemUser, systemErr = s.userRepo.GetByEmailSystem(ctx, req.Username)
			if systemErr == nil && systemUser != nil && systemUser.PrincipalType == models.PrincipalTypeSystem {
				user = systemUser
			}
		}
		
		if user == nil {
			return nil, fmt.Errorf("invalid credentials")
		}
	}

	// Check if user is active
	if !user.IsActive() {
		return nil, fmt.Errorf("user account is not active")
	}

	// Get credentials
	cred, err := s.credentialRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if account is locked
	if cred.IsLocked() {
		return nil, fmt.Errorf("account is locked due to too many failed login attempts")
	}

	// Verify password
	valid, err := s.passwordHasher.Verify(req.Password, cred.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to verify password: %w", err)
	}

	if !valid {
		// Increment failed attempts
		cred.IncrementFailedAttempts()
		if err := s.credentialRepo.Update(ctx, cred); err != nil {
			// Log error but continue
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("invalid credentials")
	}

	// Reset failed attempts on successful login
	cred.ResetFailedAttempts()
	if err := s.credentialRepo.Update(ctx, cred); err != nil {
		// Log error but continue with login
		// The credential update failure shouldn't block login
		// Error is intentionally ignored to not block successful authentication
		_ = err
	}

	// Check if MFA is required
	// MFA is required if:
	// 1. User has MFA enabled (user.MFAEnabled), OR
	// 2. Tenant requires MFA for all users (tenant settings MFARequired)
	mfaRequired := false
	
	// Check user-level MFA
	if user.MFAEnabled {
		mfaRequired = true
	} else if user.TenantID != nil {
		// Check tenant-level MFA requirement
		tenantSettings, err := s.tenantSettingsRepo.GetByTenantID(ctx, *user.TenantID)
		if err == nil && tenantSettings != nil && tenantSettings.MFARequired {
			mfaRequired = true
		}
	}
	
	if mfaRequired {
		// MFA is required - client should call /api/v1/mfa/challenge endpoint
		// with username and password to get a challenge session
		// Note: We don't return tokens yet, user must complete MFA verification
		return &LoginResponse{
			MFARequired: true,
			// Client should call MFA challenge endpoint with username/password
		}, nil
	}

	// If login_challenge is provided, use OAuth2 flow
	if req.LoginChallenge != nil {
		return s.handleOAuth2Login(ctx, *req.LoginChallenge, user)
	}

	// Direct token issuance (simplified flow)
	// For SYSTEM users, tenantID is nil
	var tenantID uuid.UUID
	if user.TenantID != nil {
		tenantID = *user.TenantID
	}
	return s.issueDirectTokens(ctx, user, tenantID, req.RememberMe)
}

// handleOAuth2Login handles OAuth2 login flow with Hydra
func (s *Service) handleOAuth2Login(ctx context.Context, challenge string, user *models.User) (*LoginResponse, error) {
	// Verify login request exists in Hydra (we don't need the full request for now)
	if _, err := s.hydraClient.GetLoginRequest(ctx, challenge); err != nil {
		return nil, fmt.Errorf("failed to get login request: %w", err)
	}

	// Build claims from user, roles, and permissions
	claimsObj, err := s.claimsBuilder.BuildClaims(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to build claims: %w", err)
	}

	// Convert claims to map for Hydra
	claims := map[string]interface{}{
		"sub":        claimsObj.Subject,
		"tenant_id":  claimsObj.TenantID,
		"email":      claimsObj.Email,
		"username":   claimsObj.Username,
		"roles":      claimsObj.Roles,
		"permissions": claimsObj.Permissions,
		"scope":      claimsObj.Scope,
	}

	// Accept login in Hydra
	acceptResp, err := s.hydraClient.AcceptLoginRequest(ctx, challenge, &hydra.AcceptLoginRequest{
		Subject:     user.ID.String(),
		Context:     claims,
		Remember:    true,
		RememberFor: 3600,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to accept login: %w", err)
	}

	return &LoginResponse{
		RedirectTo: acceptResp.RedirectTo,
	}, nil
}

