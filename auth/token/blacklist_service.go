package token

import (
	"context"
	"fmt"
	"time"

	"github.com/arauth-identity/iam/internal/cache"
	"go.uber.org/zap"
)

// BlacklistService handles token revocation and checking
type BlacklistService struct {
	cache  *cache.Cache
	logger *zap.Logger
}

// NewBlacklistService creates a new BlacklistService
func NewBlacklistService(c *cache.Cache, logger *zap.Logger) *BlacklistService {
	return &BlacklistService{
		cache:  c,
		logger: logger,
	}
}

// RevokeToken adds a token's JTI to the blacklist with an expiration
func (s *BlacklistService) RevokeToken(ctx context.Context, jti string, expiration time.Duration) error {
	key := fmt.Sprintf("blacklist:token:%s", jti)

	// Store "revoked" as value. The cache wrapper handles JSON marshaling.
	err := s.cache.Set(ctx, key, "revoked", expiration)
	if err != nil {
		s.logger.Error("failed to revoke token in redis",
			zap.String("jti", jti),
			zap.Error(err),
		)
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	s.logger.Info("token revoked", zap.String("jti", jti), zap.Duration("ttl", expiration))
	return nil
}

// IsRevoked checks if a token's JTI is in the blacklist
func (s *BlacklistService) IsRevoked(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("blacklist:token:%s", jti)

	// Use Exists method provided by cache wrapper
	exists, err := s.cache.Exists(ctx, key)
	if err != nil {
		// Redis Failure Mode: FAIL CLOSED (return error)
		s.logger.Error("failed to check token revocation status",
			zap.String("jti", jti),
			zap.Error(err),
		)
		return false, fmt.Errorf("failed to check revocation status: %w", err)
	}

	return exists, nil
}
