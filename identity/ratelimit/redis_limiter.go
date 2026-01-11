package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisLimiter implements the Limiter interface using Redis for storage
type RedisLimiter struct {
	client *redis.Client
	config *Config
}

// NewRedisLimiter creates a new Redis-backed rate limiter
func NewRedisLimiter(client *redis.Client, config *Config) *RedisLimiter {
	if config == nil {
		config = DefaultConfig()
	}
	return &RedisLimiter{
		client: client,
		config: config,
	}
}

// redisKey generates a Redis key for rate limiting
func (l *RedisLimiter) redisKey(limitType LimitType, identifier string, category EndpointCategory) string {
	if category == "" {
		category = CategoryGeneral
	}
	// Format: ratelimit:{type}:{identifier}:{category}:{window}
	window := time.Now().Unix() / int64(l.config.WindowDuration.Seconds())
	return fmt.Sprintf("ratelimit:%s:%s:%s:%d", limitType, identifier, category, window)
}

// getLimit returns the appropriate limit for a given category and limit type
func (l *RedisLimiter) getLimit(limitType LimitType, category EndpointCategory) (limit int, burst int) {
	switch limitType {
	case LimitTypeUser:
		switch category {
		case CategoryAuth:
			return l.config.AuthRequestsPerMinute, l.config.AuthBurstSize
		case CategorySensitive:
			return l.config.SensitiveRequestsPerMinute, l.config.SensitiveBurstSize
		default:
			return l.config.UserRequestsPerMinute, l.config.UserBurstSize
		}
	case LimitTypeClient:
		return l.config.ClientRequestsPerMinute, l.config.ClientBurstSize
	case LimitTypeIP:
		if category == CategoryAdmin || category == CategorySensitive {
			return l.config.AdminIPRequestsPerMinute, l.config.AdminIPBurstSize
		}
		return l.config.UserRequestsPerMinute, l.config.UserBurstSize
	default:
		return l.config.UserRequestsPerMinute, l.config.UserBurstSize
	}
}

// checkLimit is the core rate limiting logic using sliding window
func (l *RedisLimiter) checkLimit(ctx context.Context, limitType LimitType, identifier string, category EndpointCategory) error {
	key := l.redisKey(limitType, identifier, category)
	limit, burst := l.getLimit(limitType, category)

	// Use Redis INCR for atomic increment
	count, err := l.client.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to increment rate limit counter: %w", err)
	}

	// Set expiry on first request in window
	if count == 1 {
		l.client.Expire(ctx, key, l.config.WindowDuration)
	}

	// Check if limit exceeded (allow burst)
	if count > int64(limit+burst) {
		ttl, _ := l.client.TTL(ctx, key).Result()
		return &RateLimitError{
			LimitType:    limitType,
			Identifier:   identifier,
			Limit:        limit,
			WindowStart:  time.Now().Add(-l.config.WindowDuration + ttl),
			RetryAfter:   ttl,
			CurrentCount: int(count),
		}
	}

	return nil
}

// getUsage returns current usage for a key
func (l *RedisLimiter) getUsage(ctx context.Context, limitType LimitType, identifier string, category EndpointCategory) (current int, limit int, err error) {
	key := l.redisKey(limitType, identifier, category)
	limitVal, _ := l.getLimit(limitType, category)

	countStr, err := l.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, limitVal, nil
	}
	if err != nil {
		return 0, limitVal, fmt.Errorf("failed to get rate limit usage: %w", err)
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, limitVal, fmt.Errorf("invalid count value: %w", err)
	}

	return count, limitVal, nil
}

// resetLimit manually resets a rate limit counter
func (l *RedisLimiter) resetLimit(ctx context.Context, limitType LimitType, identifier string, category EndpointCategory) error {
	key := l.redisKey(limitType, identifier, category)
	return l.client.Del(ctx, key).Err()
}

// CheckUserLimit implements Limiter.CheckUserLimit
func (l *RedisLimiter) CheckUserLimit(ctx context.Context, userID string, category EndpointCategory) error {
	return l.checkLimit(ctx, LimitTypeUser, userID, category)
}

// CheckClientLimit implements Limiter.CheckClientLimit
func (l *RedisLimiter) CheckClientLimit(ctx context.Context, clientID string) error {
	return l.checkLimit(ctx, LimitTypeClient, clientID, CategoryGeneral)
}

// CheckIPLimit implements Limiter.CheckIPLimit
func (l *RedisLimiter) CheckIPLimit(ctx context.Context, ip string, category EndpointCategory) error {
	return l.checkLimit(ctx, LimitTypeIP, ip, category)
}

// GetUserUsage implements Limiter.GetUserUsage
func (l *RedisLimiter) GetUserUsage(ctx context.Context, userID string, category EndpointCategory) (current int, limit int, err error) {
	return l.getUsage(ctx, LimitTypeUser, userID, category)
}

// GetClientUsage implements Limiter.GetClientUsage
func (l *RedisLimiter) GetClientUsage(ctx context.Context, clientID string) (current int, limit int, err error) {
	return l.getUsage(ctx, LimitTypeClient, clientID, CategoryGeneral)
}

// GetIPUsage implements Limiter.GetIPUsage
func (l *RedisLimiter) GetIPUsage(ctx context.Context, ip string, category EndpointCategory) (current int, limit int, err error) {
	return l.getUsage(ctx, LimitTypeIP, ip, category)
}

// ResetUserLimit implements Limiter.ResetUserLimit
func (l *RedisLimiter) ResetUserLimit(ctx context.Context, userID string, category EndpointCategory) error {
	return l.resetLimit(ctx, LimitTypeUser, userID, category)
}

// ResetIPLimit implements Limiter.ResetIPLimit
func (l *RedisLimiter) ResetIPLimit(ctx context.Context, ip string, category EndpointCategory) error {
	return l.resetLimit(ctx, LimitTypeIP, ip, category)
}
