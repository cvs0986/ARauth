package session

import (
	"context"

	"github.com/google/uuid"
)

// ServiceInterface defines the interface for session service operations
type ServiceInterface interface {
	// ListSessions lists all active sessions for a user within a tenant
	ListSessions(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) ([]*Session, error)

	// GetSessionByID retrieves a session by its ID
	GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*Session, error)

	// RevokeSession revokes a session by its ID
	RevokeSession(ctx context.Context, sessionID uuid.UUID, reason string) error
}
