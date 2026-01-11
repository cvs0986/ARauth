package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cleanup := func() {
		client.Close()
		mr.Close()
	}

	return client, cleanup
}

func TestRedisLimiter_CheckUserLimit(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	config := &Config{
		UserRequestsPerMinute: 5,
		UserBurstSize:         2,
		WindowDuration:        time.Minute,
	}

	limiter := NewRedisLimiter(client, config)
	ctx := context.Background()
	userID := "user-123"

	// First 7 requests should succeed (5 + 2 burst)
	for i := 0; i < 7; i++ {
		err := limiter.CheckUserLimit(ctx, userID, CategoryGeneral)
		assert.NoError(t, err, "request %d should succeed", i+1)
	}

	// 8th request should fail
	err := limiter.CheckUserLimit(ctx, userID, CategoryGeneral)
	assert.Error(t, err)

	var rateLimitErr *RateLimitError
	assert.ErrorAs(t, err, &rateLimitErr)
	assert.Equal(t, LimitTypeUser, rateLimitErr.LimitType)
	assert.Equal(t, userID, rateLimitErr.Identifier)
	assert.Equal(t, 5, rateLimitErr.Limit)
	assert.Equal(t, 8, rateLimitErr.CurrentCount)
}

func TestRedisLimiter_CheckClientLimit(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	config := &Config{
		ClientRequestsPerMinute: 10,
		ClientBurstSize:         3,
		WindowDuration:          time.Minute,
	}

	limiter := NewRedisLimiter(client, config)
	ctx := context.Background()
	clientID := "client-456"

	// First 13 requests should succeed (10 + 3 burst)
	for i := 0; i < 13; i++ {
		err := limiter.CheckClientLimit(ctx, clientID)
		assert.NoError(t, err, "request %d should succeed", i+1)
	}

	// 14th request should fail
	err := limiter.CheckClientLimit(ctx, clientID)
	assert.Error(t, err)

	var rateLimitErr *RateLimitError
	assert.ErrorAs(t, err, &rateLimitErr)
	assert.Equal(t, LimitTypeClient, rateLimitErr.LimitType)
}

func TestRedisLimiter_CheckIPLimit(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	config := &Config{
		AdminIPRequestsPerMinute: 3,
		AdminIPBurstSize:         1,
		WindowDuration:           time.Minute,
	}

	limiter := NewRedisLimiter(client, config)
	ctx := context.Background()
	ip := "192.168.1.100"

	// First 4 requests should succeed (3 + 1 burst)
	for i := 0; i < 4; i++ {
		err := limiter.CheckIPLimit(ctx, ip, CategoryAdmin)
		assert.NoError(t, err, "request %d should succeed", i+1)
	}

	// 5th request should fail
	err := limiter.CheckIPLimit(ctx, ip, CategoryAdmin)
	assert.Error(t, err)

	var rateLimitErr *RateLimitError
	assert.ErrorAs(t, err, &rateLimitErr)
	assert.Equal(t, LimitTypeIP, rateLimitErr.LimitType)
	assert.Equal(t, ip, rateLimitErr.Identifier)
}

func TestRedisLimiter_CategoryLimits(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	config := &Config{
		UserRequestsPerMinute:      60,
		UserBurstSize:              10,
		AuthRequestsPerMinute:      20,
		AuthBurstSize:              3,
		SensitiveRequestsPerMinute: 10,
		SensitiveBurstSize:         2,
		WindowDuration:             time.Minute,
	}

	limiter := NewRedisLimiter(client, config)
	ctx := context.Background()
	userID := "user-789"

	// Test general category (60 + 10 burst = 70)
	for i := 0; i < 70; i++ {
		err := limiter.CheckUserLimit(ctx, userID, CategoryGeneral)
		assert.NoError(t, err)
	}
	err := limiter.CheckUserLimit(ctx, userID, CategoryGeneral)
	assert.Error(t, err)

	// Test auth category (20 + 3 burst = 23) - different user
	userID2 := "user-auth"
	for i := 0; i < 23; i++ {
		err := limiter.CheckUserLimit(ctx, userID2, CategoryAuth)
		assert.NoError(t, err)
	}
	err = limiter.CheckUserLimit(ctx, userID2, CategoryAuth)
	assert.Error(t, err)

	// Test sensitive category (10 + 2 burst = 12) - different user
	userID3 := "user-sensitive"
	for i := 0; i < 12; i++ {
		err := limiter.CheckUserLimit(ctx, userID3, CategorySensitive)
		assert.NoError(t, err)
	}
	err = limiter.CheckUserLimit(ctx, userID3, CategorySensitive)
	assert.Error(t, err)
}

func TestRedisLimiter_GetUsage(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	config := &Config{
		UserRequestsPerMinute: 10,
		UserBurstSize:         2,
		WindowDuration:        time.Minute,
	}

	limiter := NewRedisLimiter(client, config)
	ctx := context.Background()
	userID := "user-usage"

	// Initial usage should be 0
	current, limit, err := limiter.GetUserUsage(ctx, userID, CategoryGeneral)
	assert.NoError(t, err)
	assert.Equal(t, 0, current)
	assert.Equal(t, 10, limit)

	// Make 5 requests
	for i := 0; i < 5; i++ {
		limiter.CheckUserLimit(ctx, userID, CategoryGeneral)
	}

	// Usage should be 5
	current, limit, err = limiter.GetUserUsage(ctx, userID, CategoryGeneral)
	assert.NoError(t, err)
	assert.Equal(t, 5, current)
	assert.Equal(t, 10, limit)
}

func TestRedisLimiter_ResetLimit(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	config := &Config{
		UserRequestsPerMinute: 3,
		UserBurstSize:         1,
		WindowDuration:        time.Minute,
	}

	limiter := NewRedisLimiter(client, config)
	ctx := context.Background()
	userID := "user-reset"

	// Exhaust limit
	for i := 0; i < 4; i++ {
		limiter.CheckUserLimit(ctx, userID, CategoryGeneral)
	}

	// Should be rate limited
	err := limiter.CheckUserLimit(ctx, userID, CategoryGeneral)
	assert.Error(t, err)

	// Reset limit
	err = limiter.ResetUserLimit(ctx, userID, CategoryGeneral)
	assert.NoError(t, err)

	// Should succeed now
	err = limiter.CheckUserLimit(ctx, userID, CategoryGeneral)
	assert.NoError(t, err)
}

func TestRedisLimiter_IsolationBetweenIdentifiers(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	config := &Config{
		UserRequestsPerMinute: 2,
		UserBurstSize:         0,
		WindowDuration:        time.Minute,
	}

	limiter := NewRedisLimiter(client, config)
	ctx := context.Background()

	// User 1 exhausts limit
	for i := 0; i < 2; i++ {
		err := limiter.CheckUserLimit(ctx, "user-1", CategoryGeneral)
		assert.NoError(t, err)
	}
	err := limiter.CheckUserLimit(ctx, "user-1", CategoryGeneral)
	assert.Error(t, err)

	// User 2 should still have full quota
	err = limiter.CheckUserLimit(ctx, "user-2", CategoryGeneral)
	assert.NoError(t, err)
	err = limiter.CheckUserLimit(ctx, "user-2", CategoryGeneral)
	assert.NoError(t, err)
	err = limiter.CheckUserLimit(ctx, "user-2", CategoryGeneral)
	assert.Error(t, err)
}

func TestRedisLimiter_IsolationBetweenCategories(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	config := &Config{
		UserRequestsPerMinute: 5,
		UserBurstSize:         0,
		AuthRequestsPerMinute: 2,
		AuthBurstSize:         0,
		WindowDuration:        time.Minute,
	}

	limiter := NewRedisLimiter(client, config)
	ctx := context.Background()
	userID := "user-categories"

	// Exhaust auth category
	for i := 0; i < 2; i++ {
		err := limiter.CheckUserLimit(ctx, userID, CategoryAuth)
		assert.NoError(t, err)
	}
	err := limiter.CheckUserLimit(ctx, userID, CategoryAuth)
	assert.Error(t, err)

	// General category should still have full quota
	for i := 0; i < 5; i++ {
		err := limiter.CheckUserLimit(ctx, userID, CategoryGeneral)
		assert.NoError(t, err)
	}
	err = limiter.CheckUserLimit(ctx, userID, CategoryGeneral)
	assert.Error(t, err)
}

func TestRedisLimiter_DefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, 60, config.UserRequestsPerMinute)
	assert.Equal(t, 10, config.UserBurstSize)
	assert.Equal(t, 100, config.ClientRequestsPerMinute)
	assert.Equal(t, 20, config.ClientBurstSize)
	assert.Equal(t, 30, config.AdminIPRequestsPerMinute)
	assert.Equal(t, 5, config.AdminIPBurstSize)
	assert.Equal(t, 20, config.AuthRequestsPerMinute)
	assert.Equal(t, 3, config.AuthBurstSize)
	assert.Equal(t, 10, config.SensitiveRequestsPerMinute)
	assert.Equal(t, 2, config.SensitiveBurstSize)
	assert.Equal(t, time.Minute, config.WindowDuration)
}
