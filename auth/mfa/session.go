package mfa

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/internal/cache"
)

// SessionManager manages MFA sessions
type SessionManager struct {
	cache *cache.Cache
	ttl   time.Duration
}

// NewSessionManager creates a new MFA session manager
func NewSessionManager(cacheClient *cache.Cache) *SessionManager {
	return &SessionManager{
		cache: cacheClient,
		ttl:   5 * time.Minute, // MFA sessions expire after 5 minutes
	}
}

// MFASession represents an MFA challenge session
type MFASession struct {
	SessionID  string    `json:"session_id"`
	UserID     uuid.UUID `json:"user_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	Attempts   int       `json:"attempts"`
	MaxAttempts int      `json:"max_attempts"`
}

// CreateSession creates a new MFA session
func (sm *SessionManager) CreateSession(ctx context.Context, userID, tenantID uuid.UUID) (string, error) {
	sessionID := uuid.New().String()
	now := time.Now()

	session := &MFASession{
		SessionID:  sessionID,
		UserID:     userID,
		TenantID:   tenantID,
		CreatedAt:  now,
		ExpiresAt:  now.Add(sm.ttl),
		Attempts:   0,
		MaxAttempts: 5,
	}

	key := fmt.Sprintf("mfa:session:%s", sessionID)
	if err := sm.cache.Set(ctx, key, session, sm.ttl); err != nil {
		return "", fmt.Errorf("failed to create MFA session: %w", err)
	}

	return sessionID, nil
}

// GetSession retrieves an MFA session
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*MFASession, error) {
	key := fmt.Sprintf("mfa:session:%s", sessionID)

	var session MFASession
	if err := sm.cache.Get(ctx, key, &session); err != nil {
		return nil, fmt.Errorf("session not found or expired: %w", err)
	}

	// Check if session expired
	if time.Now().After(session.ExpiresAt) {
		_ = sm.DeleteSession(ctx, sessionID) // Ignore error on cleanup
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

// IncrementAttempts increments the attempt counter for a session
func (sm *SessionManager) IncrementAttempts(ctx context.Context, sessionID string) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.Attempts++
	if session.Attempts >= session.MaxAttempts {
		_ = sm.DeleteSession(ctx, sessionID) // Ignore error on cleanup
		return fmt.Errorf("maximum attempts exceeded")
	}

	key := fmt.Sprintf("mfa:session:%s", sessionID)
	remainingTTL := time.Until(session.ExpiresAt)
	if err := sm.cache.Set(ctx, key, session, remainingTTL); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// DeleteSession deletes an MFA session
func (sm *SessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("mfa:session:%s", sessionID)
	return sm.cache.Delete(ctx, key)
}

// VerifySession verifies a session is valid and not exceeded attempts
func (sm *SessionManager) VerifySession(ctx context.Context, sessionID string) (*MFASession, error) {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.Attempts >= session.MaxAttempts {
		_ = sm.DeleteSession(ctx, sessionID) // Ignore error on cleanup
		return nil, fmt.Errorf("maximum attempts exceeded")
	}

	return session, nil
}

