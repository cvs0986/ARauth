package session

import (
	"context"
	"fmt"

	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/google/uuid"
)

// Service implements the session service
type Service struct {
	refreshTokenRepo interfaces.RefreshTokenRepository
	userRepo         interfaces.UserRepository
}

// NewService creates a new session service
func NewService(refreshTokenRepo interfaces.RefreshTokenRepository, userRepo interfaces.UserRepository) *Service {
	return &Service{
		refreshTokenRepo: refreshTokenRepo,
		userRepo:         userRepo,
	}
}

// ListSessions lists all active sessions for a user within a tenant
func (s *Service) ListSessions(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]*Session, error) {
	// Get all refresh tokens for the user
	tokens, err := s.refreshTokenRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh tokens: %w", err)
	}

	// Get user info for username
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Filter by tenant and map to sessions
	sessions := make([]*Session, 0)
	for _, token := range tokens {
		// Tenant isolation: only include tokens for the specified tenant
		if token.TenantID != tenantID {
			continue
		}

		// Skip revoked tokens
		if token.RevokedAt != nil {
			continue
		}

		session := s.mapTokenToSession(token, user.Username)
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// GetSessionByID retrieves a session by its ID
func (s *Service) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*Session, error) {
	// Get the refresh token by ID
	// Note: RefreshTokenRepository doesn't have GetByID, so we need to get by user and filter
	// This is a limitation we'll work around for now
	// In production, you'd add GetByID to the repository interface

	// For now, we'll return an error indicating this needs implementation
	// The handler will need to verify ownership differently
	return nil, fmt.Errorf("GetSessionByID not yet implemented - use ListSessions and filter")
}

// RevokeSession revokes a session by its ID
func (s *Service) RevokeSession(ctx context.Context, sessionID uuid.UUID, reason string) error {
	// Revoke the refresh token
	if err := s.refreshTokenRepo.Revoke(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	return nil
}

// mapTokenToSession maps a refresh token to a session model
func (s *Service) mapTokenToSession(token *interfaces.RefreshToken, username string) *Session {
	session := &Session{
		ID:              token.ID,
		UserID:          token.UserID,
		Username:        username,
		CreatedAt:       token.CreatedAt,
		ExpiresAt:       token.ExpiresAt,
		RememberMe:      token.RememberMe,
		MFAVerified:     token.MFAVerified,
		IsImpersonation: false, // Default to false
		ImpersonatorID:  nil,
	}

	// TODO: Detect impersonation sessions
	// This requires token metadata which isn't currently stored in RefreshToken
	// For now, all sessions are marked as non-impersonated
	// In production, you'd check token metadata for impersonator_id

	return session
}
