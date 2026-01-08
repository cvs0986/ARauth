package login

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/auth/hydra"
	"github.com/arauth-identity/iam/auth/token"
	"github.com/arauth-identity/iam/identity/capability"
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
	capabilityService   capability.ServiceInterface
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
	capabilityService capability.ServiceInterface,
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
		capabilityService: capabilityService,
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
	UserID           string `json:"user_id,omitempty"`   // Return user ID when MFA is required
	TenantID         string `json:"tenant_id,omitempty"` // Return tenant ID when MFA is required
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

	// Check if password authentication is allowed (for tenant users)
	// SYSTEM users always allowed (they don't have tenant restrictions)
	if user.TenantID != nil {
		// Check if password auth capability is allowed for tenant
		// Note: We assume password auth is always supported at system level
		// If we add a "password" capability, we would check it here
		// For now, password auth is always allowed if tenant exists
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

	// Check if MFA is required and allowed
	// MFA is required if:
	// 1. User has MFA enabled (user.MFAEnabled), OR
	// 2. Tenant requires MFA for all users (tenant settings MFARequired)
	// But MFA must also be:
	// 3. Supported by system
	// 4. Allowed for tenant
	// 5. Enabled by tenant
	mfaRequired := false
	
	// First check if MFA/TOTP is allowed and enabled via capability model
	var mfaAllowed bool
	var mfaEnabled bool
	if user.TenantID != nil {
		// Check capability model for tenant users
		eval, err := s.capabilityService.EvaluateCapability(ctx, *user.TenantID, user.ID, models.CapabilityKeyMFA)
		if err == nil && eval != nil {
			mfaAllowed = eval.TenantAllowed
			mfaEnabled = eval.TenantEnabled
		}
	} else {
		// For SYSTEM users, check if MFA is supported at system level
		supported, err := s.capabilityService.IsCapabilitySupported(ctx, models.CapabilityKeyMFA)
		if err == nil {
			mfaAllowed = supported
			mfaEnabled = supported // System users can use MFA if supported
		}
	}
	
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
	
	// If MFA is required but not allowed/enabled, return error
	if mfaRequired && (!mfaAllowed || !mfaEnabled) {
		return nil, fmt.Errorf("MFA is required but not available for this tenant")
	}
	
	if mfaRequired {
		// MFA is required - return user info so client can call MFA challenge
		// The client will need to call /api/v1/mfa/challenge with user_id and tenant_id
		var tenantIDStr string
		if user.TenantID != nil {
			tenantIDStr = user.TenantID.String()
		}
		return &LoginResponse{
			MFARequired: true,
			UserID:     user.ID.String(),
			TenantID:   tenantIDStr,
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
	// Check if OAuth2/OIDC is allowed for tenant
	if user.TenantID != nil {
		eval, err := s.capabilityService.EvaluateCapability(ctx, *user.TenantID, user.ID, models.CapabilityKeyOIDC)
		if err == nil && eval != nil && !eval.CanUse {
			return nil, fmt.Errorf("OAuth2/OIDC is not available for this tenant: %s", eval.Reason)
		}
	} else {
		// For SYSTEM users, check if OIDC is supported
		supported, err := s.capabilityService.IsCapabilitySupported(ctx, models.CapabilityKeyOIDC)
		if err != nil || !supported {
			return nil, fmt.Errorf("OAuth2/OIDC is not supported")
		}
	}

	// Get login request from Hydra to validate scopes
	loginReq, err := s.hydraClient.GetLoginRequest(ctx, challenge)
	if err != nil {
		return nil, fmt.Errorf("failed to get login request: %w", err)
	}

	// Validate requested scopes against allowed scope namespaces
	if user.TenantID != nil && len(loginReq.RequestedScope) > 0 {
		// Get tenant capability for allowed scope namespaces
		tenantCap, err := s.capabilityService.GetSystemCapability(ctx, models.CapabilityKeyAllowedScopeNamespaces)
		if err == nil && tenantCap != nil {
			// Get system default allowed namespaces
			allowedNamespaces, err := tenantCap.GetDefaultValue()
			if err == nil {
				if namespaces, ok := allowedNamespaces["value"].([]interface{}); ok {
					allowedNamespaceMap := make(map[string]bool)
					for _, ns := range namespaces {
						if nsStr, ok := ns.(string); ok {
							allowedNamespaceMap[nsStr] = true
						}
					}
					
					// TODO: Check tenant-specific allowed namespaces in Phase 3
					// For now, use system defaults
					
					// Validate each requested scope
					for _, requestedScope := range loginReq.RequestedScope {
						// Extract namespace from scope (e.g., "users:read" -> "users")
						namespace := requestedScope
						for i, char := range requestedScope {
							if char == ':' {
								namespace = requestedScope[:i]
								break
							}
						}
						
						// Standard OIDC scopes are always allowed
						standardScopes := map[string]bool{
							"openid":        true,
							"profile":       true,
							"email":         true,
							"offline_access": true,
						}
						
						// Check if namespace is allowed (or if it's a standard OIDC scope)
						if !standardScopes[namespace] && !standardScopes[requestedScope] {
							if !allowedNamespaceMap[namespace] {
								return nil, fmt.Errorf("scope namespace '%s' is not allowed for this tenant", namespace)
							}
						}
					}
				}
			}
		}
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

