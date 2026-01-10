package token

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/arauth-identity/iam/internal/cache"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupBlacklistService(t *testing.T) (*BlacklistService, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	redisCache := cache.NewCache(rdb)
	logger := zap.NewNop()

	return NewBlacklistService(redisCache, logger), mr
}

func TestBlacklistService_RevokeToken(t *testing.T) {
	service, mr := setupBlacklistService(t)
	defer mr.Close()

	ctx := context.Background()
	jti := "test-jti-123"
	ttl := time.Hour

	// Test Revoke
	err := service.RevokeToken(ctx, jti, ttl)
	assert.NoError(t, err)

	// Verify in Miniredis
	// Key format: blacklist:token:<jti>
	key := "blacklist:token:" + jti
	assert.True(t, mr.Exists(key))

	// Check TTL
	// Miniredis TTL is precise enough for check
	assert.Greater(t, mr.TTL(key), time.Duration(0))
}

func TestBlacklistService_IsRevoked(t *testing.T) {
	service, mr := setupBlacklistService(t)
	defer mr.Close()

	ctx := context.Background()
	jti := "test-jti-revoked"

	// Initially not revoked
	revoked, err := service.IsRevoked(ctx, jti)
	assert.NoError(t, err)
	assert.False(t, revoked)

	// Revoke it
	err = service.RevokeToken(ctx, jti, time.Hour)
	assert.NoError(t, err)

	// Check again
	revoked, err = service.IsRevoked(ctx, jti)
	assert.NoError(t, err)
	assert.True(t, revoked)
}

func TestBlacklistService_RedisFailure(t *testing.T) {
	// This test is hard to simulate with Miniredis because closing it kills the connection entirely.
	// But we can try checking behavior when connection fails.

	mr, err := miniredis.Run()
	assert.NoError(t, err)

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	mr.Close() // Close immediately to simulate failure

	redisCache := cache.NewCache(rdb)
	logger := zap.NewNop()
	service := NewBlacklistService(redisCache, logger)

	ctx := context.Background()
	jti := "test-jti-failure"

	// Expect Error on IsRevoked (Fail Closed)
	revoked, err := service.IsRevoked(ctx, jti)
	assert.Error(t, err)
	assert.False(t, revoked) // Should fail closed (err returned), bool val implies safe-ish/false usually but err matters.
	// Wait, safe default?
	// Logic: return false, error.
	// Middleware checks: if err != nil -> 401. So it works.
}
