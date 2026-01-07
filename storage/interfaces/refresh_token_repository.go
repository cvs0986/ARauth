package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token
type RefreshToken struct {
	ID         uuid.UUID  `db:"id"`
	UserID     uuid.UUID  `db:"user_id"`
	TenantID   uuid.UUID  `db:"tenant_id"`
	TokenHash  string     `db:"token_hash"`
	ExpiresAt  time.Time  `db:"expires_at"`
	RevokedAt  *time.Time `db:"revoked_at"`
	RememberMe bool       `db:"remember_me"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
}

// RefreshTokenRepository defines operations for refresh tokens
type RefreshTokenRepository interface {
	// Create creates a new refresh token
	Create(ctx context.Context, token *RefreshToken) error

	// GetByTokenHash retrieves a refresh token by its hash
	GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)

	// GetByUserID retrieves all active refresh tokens for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*RefreshToken, error)

	// Revoke revokes a refresh token
	Revoke(ctx context.Context, tokenID uuid.UUID) error

	// RevokeByTokenHash revokes a refresh token by its hash
	RevokeByTokenHash(ctx context.Context, tokenHash string) error

	// RevokeAllForUser revokes all refresh tokens for a user
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error

	// DeleteExpired deletes expired refresh tokens
	DeleteExpired(ctx context.Context) error
}

