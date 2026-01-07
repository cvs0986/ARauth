package login

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/auth/hydra"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/security/password"
	"github.com/nuage-identity/iam/storage/interfaces"
)

// Service provides login functionality
type Service struct {
	userRepo       interfaces.UserRepository
	credentialRepo interfaces.CredentialRepository
	hydraClient    *hydra.Client
	passwordHasher *password.Hasher
}

// NewService creates a new login service
func NewService(
	userRepo interfaces.UserRepository,
	credentialRepo interfaces.CredentialRepository,
	hydraClient *hydra.Client,
) *Service {
	return &Service{
		userRepo:       userRepo,
		credentialRepo: credentialRepo,
		hydraClient:    hydraClient,
		passwordHasher: password.NewHasher(),
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username      string    `json:"username" binding:"required"`
	Password      string    `json:"password" binding:"required"`
	TenantID      uuid.UUID `json:"tenant_id" binding:"required"`
	LoginChallenge *string  `json:"login_challenge,omitempty"` // For OAuth2 flow
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	MFARequired  bool   `json:"mfa_required"`
	MFASessionID string `json:"mfa_session_id,omitempty"`
	RedirectTo   string `json:"redirect_to,omitempty"` // For OAuth2 flow
}

// Login authenticates a user and returns tokens
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Validate tenant ID is provided
	if req.TenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// Get user by username (tenant-scoped)
	user, err := s.userRepo.GetByUsername(ctx, req.Username, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
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
		s.credentialRepo.Update(ctx, cred)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Reset failed attempts on successful login
	cred.ResetFailedAttempts()
	s.credentialRepo.Update(ctx, cred)

	// Check if MFA is required
	if user.MFAEnabled {
		// MFA is required - client should call /api/v1/mfa/challenge endpoint
		// with user_id and tenant_id to get a session
		return &LoginResponse{
			MFARequired: true,
			// Client should call MFA challenge endpoint to get session
		}, nil
	}

	// If login_challenge is provided, use OAuth2 flow
	if req.LoginChallenge != nil {
		return s.handleOAuth2Login(ctx, *req.LoginChallenge, user)
	}

	// TODO: Direct token issuance (simplified flow)
	// For now, return success but tokens need to be issued via Hydra
	return &LoginResponse{
		TokenType: "Bearer",
		// Tokens will be issued via Hydra
	}, nil
}

// handleOAuth2Login handles OAuth2 login flow with Hydra
func (s *Service) handleOAuth2Login(ctx context.Context, challenge string, user *models.User) (*LoginResponse, error) {
	// Verify login request exists in Hydra (we don't need the full request for now)
	if _, err := s.hydraClient.GetLoginRequest(ctx, challenge); err != nil {
		return nil, fmt.Errorf("failed to get login request: %w", err)
	}

	// Build claims (simplified for now)
	claims := map[string]interface{}{
		"sub":    user.ID.String(),
		"tenant": user.TenantID.String(),
		"email":  user.Email,
		"username": user.Username,
	}

	// Accept login in Hydra
	acceptResp, err := s.hydraClient.AcceptLoginRequest(ctx, challenge, &hydra.AcceptLoginRequest{
		Subject: user.ID.String(),
		Context: claims,
		Remember: true,
		RememberFor: 3600,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to accept login: %w", err)
	}

	return &LoginResponse{
		RedirectTo: acceptResp.RedirectTo,
	}, nil
}

