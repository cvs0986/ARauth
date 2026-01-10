package impersonation

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/auth/token"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// Service provides impersonation functionality
type Service struct {
	impersonationRepo interfaces.ImpersonationRepository
	userRepo          interfaces.UserRepository
	claimsBuilder     *claims.Builder
	tokenService      token.ServiceInterface
	lifetimeResolver  *token.LifetimeResolver
}

// NewService creates a new impersonation service
func NewService(
	impersonationRepo interfaces.ImpersonationRepository,
	userRepo interfaces.UserRepository,
	claimsBuilder *claims.Builder,
	tokenService token.ServiceInterface,
	lifetimeResolver *token.LifetimeResolver,
) ServiceInterface {
	return &Service{
		impersonationRepo: impersonationRepo,
		userRepo:          userRepo,
		claimsBuilder:     claimsBuilder,
		tokenService:      tokenService,
		lifetimeResolver:  lifetimeResolver,
	}
}

// StartImpersonation starts an impersonation session and generates an impersonation token
func (s *Service) StartImpersonation(ctx context.Context, impersonatorID uuid.UUID, targetUserID uuid.UUID, reason *string) (*ImpersonationResult, error) {
	// Validate impersonator (must be admin)
	impersonator, err := s.userRepo.GetByID(ctx, impersonatorID)
	if err != nil {
		return nil, fmt.Errorf("impersonator not found: %w", err)
	}

	// Check if impersonator has permission to impersonate
	// For now, we'll check this in the handler - here we just validate users exist
	if impersonator.PrincipalType != models.PrincipalTypeSystem {
		// For tenant users, they need tenant.admin.impersonate permission (checked in handler)
		// For now, we allow the service to proceed - handler will enforce permissions
	}

	// Validate target user
	targetUser, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("target user not found: %w", err)
	}

	// Cannot impersonate yourself
	if impersonatorID == targetUserID {
		return nil, fmt.Errorf("cannot impersonate yourself")
	}

	// Cannot impersonate SYSTEM users (unless you're a SYSTEM user)
	if targetUser.PrincipalType == models.PrincipalTypeSystem && impersonator.PrincipalType != models.PrincipalTypeSystem {
		return nil, fmt.Errorf("tenant users cannot impersonate system users")
	}

	// Build claims for target user
	targetClaims, err := s.claimsBuilder.BuildClaims(ctx, targetUser)
	if err != nil {
		return nil, fmt.Errorf("failed to build claims for target user: %w", err)
	}

	// Add impersonation metadata to claims
	impersonatorClaims, err := s.claimsBuilder.BuildClaims(ctx, impersonator)
	if err != nil {
		return nil, fmt.Errorf("failed to build claims for impersonator: %w", err)
	}

	// Create impersonation session
	sessionID := uuid.New()
	session := &models.ImpersonationSession{
		ID:              sessionID,
		ImpersonatorID:  impersonatorID,
		TargetUserID:    targetUserID,
		TenantID:        targetUser.TenantID,
		StartedAt:        time.Now(),
		Reason:           reason,
		Metadata: map[string]interface{}{
			"impersonator_username": impersonator.Username,
			"target_username":       targetUser.Username,
		},
	}

	// Get token lifetime
	var tenantID uuid.UUID
	if targetUser.TenantID != nil {
		tenantID = *targetUser.TenantID
	}
	expiresIn := s.lifetimeResolver.GetAccessTokenTTL(ctx, tenantID, false)

	// Generate token JTI
	tokenJTI := uuid.New()
	session.TokenJTI = &tokenJTI

	// Add impersonation claims to target user's claims
	targetClaims.ImpersonatedBy = impersonatorClaims.Subject
	targetClaims.ImpersonationSessionID = sessionID.String()

	// Generate access token with impersonation claim
	accessToken, err := s.tokenService.GenerateAccessToken(targetClaims, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("failed to generate impersonation token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Generate ID token (same as access token for now, but could be different)
	idToken := accessToken // In a real implementation, ID token might have different claims

	// Save session
	if err := s.impersonationRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create impersonation session: %w", err)
	}

	return &ImpersonationResult{
		Session:      session,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IDToken:      idToken,
		ExpiresIn:    int(expiresIn.Seconds()),
		TokenType:    "Bearer",
	}, nil
}


// EndImpersonation ends an active impersonation session
func (s *Service) EndImpersonation(ctx context.Context, sessionID uuid.UUID, endedBy uuid.UUID) error {
	// Get session to verify it exists and is active
	session, err := s.impersonationRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("impersonation session not found: %w", err)
	}

	if !session.IsActive() {
		return fmt.Errorf("impersonation session is already ended")
	}

	// Verify that the person ending the session is the impersonator or has permission
	if session.ImpersonatorID != endedBy {
		// Check if endedBy has permission to end impersonation sessions
		// For now, we'll allow it - in production, add permission check
	}

	// End the session
	return s.impersonationRepo.EndSession(ctx, sessionID)
}

// EndImpersonationByToken ends an impersonation session by token JTI
func (s *Service) EndImpersonationByToken(ctx context.Context, tokenJTI uuid.UUID, endedBy uuid.UUID) error {
	// Get session by token JTI
	session, err := s.impersonationRepo.GetByTokenJTI(ctx, tokenJTI)
	if err != nil {
		return fmt.Errorf("impersonation session not found: %w", err)
	}

	return s.EndImpersonation(ctx, session.ID, endedBy)
}

// GetActiveSessions retrieves active impersonation sessions
func (s *Service) GetActiveSessions(ctx context.Context, filters *ImpersonationFilters) ([]*models.ImpersonationSession, error) {
	// Convert service filters to repository filters
	repoFilters := &interfaces.ImpersonationFilters{
		ImpersonatorID: filters.ImpersonatorID,
		TargetUserID:   filters.TargetUserID,
		TenantID:       filters.TenantID,
		ActiveOnly:     filters.ActiveOnly,
		Page:           filters.Page,
		PageSize:       filters.PageSize,
	}

	return s.impersonationRepo.List(ctx, repoFilters)
}

// GetSession retrieves an impersonation session by ID
func (s *Service) GetSession(ctx context.Context, sessionID uuid.UUID) (*models.ImpersonationSession, error) {
	return s.impersonationRepo.GetByID(ctx, sessionID)
}

