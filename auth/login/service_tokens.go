package login

import (
	"context"
	"fmt"
	"time"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/google/uuid"
)

// issueDirectTokens issues access and refresh tokens directly
func (s *Service) issueDirectTokens(ctx context.Context, user *models.User, tenantID uuid.UUID, rememberMe bool) (*LoginResponse, error) {
	// Get token lifetimes
	lifetimes := s.lifetimeResolver.GetAllLifetimes(ctx, tenantID, rememberMe)

	// Build claims
	claimsObj, err := s.claimsBuilder.BuildClaims(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to build claims: %w", err)
	}

	// Set AMR claim (assuming password authentication for direct tokens)
	claimsObj.AMR = []string{"pwd"}

	// Generate access token
	accessToken, err := s.tokenService.GenerateAccessToken(claimsObj, lifetimes.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Hash refresh token for storage
	refreshTokenHash, err := s.tokenService.HashRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to hash refresh token: %w", err)
	}

	// Store refresh token
	// For SYSTEM users, tenantID is uuid.Nil (will be stored as NULL in DB)
	refreshTokenRecord := &interfaces.RefreshToken{
		UserID:     user.ID,
		TenantID:   tenantID, // uuid.Nil for SYSTEM users
		TokenHash:  refreshTokenHash,
		ExpiresAt:  time.Now().Add(lifetimes.RefreshTokenTTL),
		RememberMe: rememberMe,
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshTokenRecord); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Generate ID token (same as access token for now, can be enhanced later)
	idToken, err := s.tokenService.GenerateAccessToken(claimsObj, lifetimes.IDTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID token: %w", err)
	}

	return &LoginResponse{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken, // Return plain token to client
		IDToken:          idToken,
		TokenType:        "Bearer",
		ExpiresIn:        int(lifetimes.AccessTokenTTL.Seconds()),
		RefreshExpiresIn: int(lifetimes.RefreshTokenTTL.Seconds()),
		RememberMe:       rememberMe,
	}, nil
}
