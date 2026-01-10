package token

import (
	"context"
	"time"

	"github.com/arauth-identity/iam/auth/claims"
)

// ServiceInterface defines the interface for token service
type ServiceInterface interface {
	// GenerateAccessToken generates a JWT access token
	GenerateAccessToken(claimsObj *claims.Claims, expiresIn time.Duration) (string, error)

	// GenerateRefreshToken generates an opaque refresh token (UUID)
	GenerateRefreshToken() (string, error)

	// HashRefreshToken hashes a refresh token for storage
	HashRefreshToken(token string) (string, error)

	// VerifyRefreshToken verifies a refresh token against its hash
	VerifyRefreshToken(token, hash string) bool

	// ValidateAccessToken validates and parses an access token
	ValidateAccessToken(tokenString string) (*claims.Claims, error)

	// GetPublicKey returns the public key for JWKS endpoint
	GetPublicKey() interface{}

	// RevokeAccessToken revokes an access token
	RevokeAccessToken(ctx context.Context, tokenString string) error

	// IsAccessTokenRevoked checks if a token is revoked
	IsAccessTokenRevoked(ctx context.Context, jti string) (bool, error)
}
