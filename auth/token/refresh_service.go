package token

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/auth/claims"
	"github.com/nuage-identity/iam/storage/interfaces"
)

// RefreshService handles token refresh operations
type RefreshService struct {
	tokenService      ServiceInterface
	refreshTokenRepo interfaces.RefreshTokenRepository
	userRepo         interfaces.UserRepository
	claimsBuilder    *claims.Builder
	lifetimeResolver *LifetimeResolver
}

// NewRefreshService creates a new refresh service
func NewRefreshService(
	tokenService ServiceInterface,
	refreshTokenRepo interfaces.RefreshTokenRepository,
	userRepo interfaces.UserRepository,
	claimsBuilder *claims.Builder,
	lifetimeResolver *LifetimeResolver,
) *RefreshService {
	return &RefreshService{
		tokenService:      tokenService,
		refreshTokenRepo: refreshTokenRepo,
		userRepo:         userRepo,
		claimsBuilder:    claimsBuilder,
		lifetimeResolver: lifetimeResolver,
	}
}

// RefreshToken refreshes an access token using a refresh token
func (s *RefreshService) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error) {
	// Hash the refresh token to look it up
	refreshTokenHash, err := s.tokenService.HashRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to hash refresh token: %w", err)
	}

	// Get refresh token from database
	tokenRecord, err := s.refreshTokenRepo.GetByTokenHash(ctx, refreshTokenHash)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Check if token is revoked
	if tokenRecord.RevokedAt != nil {
		return nil, fmt.Errorf("refresh token has been revoked")
	}

	// Check if token is expired
	if time.Now().After(tokenRecord.ExpiresAt) {
		return nil, fmt.Errorf("refresh token has expired")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, tokenRecord.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is active
	if !user.IsActive() {
		return nil, fmt.Errorf("user account is not active")
	}

	// Revoke old refresh token (token rotation)
	if err := s.refreshTokenRepo.RevokeByTokenHash(ctx, refreshTokenHash); err != nil {
		// Log error but continue - token rotation is best effort
		_ = err
	}

	// Get token lifetimes
	lifetimes := s.lifetimeResolver.GetAllLifetimes(ctx, tokenRecord.TenantID, tokenRecord.RememberMe)

	// Build claims
	claimsObj, err := s.claimsBuilder.BuildClaims(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to build claims: %w", err)
	}

	// Generate new access token
	accessToken, err := s.tokenService.GenerateAccessToken(claimsObj, lifetimes.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate new refresh token
	newRefreshToken, err := s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Hash new refresh token
	newRefreshTokenHash, err := s.tokenService.HashRefreshToken(newRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to hash refresh token: %w", err)
	}

	// Store new refresh token
	newTokenRecord := &interfaces.RefreshToken{
		UserID:     user.ID,
		TenantID:   tokenRecord.TenantID,
		TokenHash:  newRefreshTokenHash,
		ExpiresAt:  time.Now().Add(lifetimes.RefreshTokenTTL),
		RememberMe: tokenRecord.RememberMe,
	}

	if err := s.refreshTokenRepo.Create(ctx, newTokenRecord); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &RefreshTokenResponse{
		AccessToken:      accessToken,
		RefreshToken:     newRefreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        int(lifetimes.AccessTokenTTL.Seconds()),
		RefreshExpiresIn: int(lifetimes.RefreshTokenTTL.Seconds()),
	}, nil
}

