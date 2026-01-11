package ratelimit

import (
	"context"
	"fmt"
	"time"
)

// LimitType defines the type of rate limit being applied
type LimitType string

const (
	LimitTypeUser   LimitType = "user"
	LimitTypeClient LimitType = "client"
	LimitTypeIP     LimitType = "ip"
)

// EndpointCategory defines the category of endpoint for tiered rate limiting
type EndpointCategory string

const (
	CategoryGeneral   EndpointCategory = "general"
	CategoryAuth      EndpointCategory = "auth"
	CategoryAdmin     EndpointCategory = "admin"
	CategorySensitive EndpointCategory = "sensitive"
)

// Config holds rate limiting configuration
type Config struct {
	// User limits
	UserRequestsPerMinute int
	UserBurstSize         int

	// Client limits
	ClientRequestsPerMinute int
	ClientBurstSize         int

	// IP limits (admin endpoints)
	AdminIPRequestsPerMinute int
	AdminIPBurstSize         int

	// Auth endpoint limits (stricter)
	AuthRequestsPerMinute int
	AuthBurstSize         int

	// Sensitive operation limits (strictest)
	SensitiveRequestsPerMinute int
	SensitiveBurstSize         int

	// Window duration for sliding window algorithm
	WindowDuration time.Duration
}

// DefaultConfig returns conservative default rate limits
func DefaultConfig() *Config {
	return &Config{
		// User limits (60 requests/min = 1 req/sec)
		UserRequestsPerMinute: 60,
		UserBurstSize:         10,

		// Client limits (100 requests/min for OAuth clients)
		ClientRequestsPerMinute: 100,
		ClientBurstSize:         20,

		// Admin IP limits (30 requests/min for admin endpoints)
		AdminIPRequestsPerMinute: 30,
		AdminIPBurstSize:         5,

		// Auth limits (stricter for login/token endpoints)
		AuthRequestsPerMinute: 20,
		AuthBurstSize:         3,

		// Sensitive limits (strictest for MFA, password reset, etc.)
		SensitiveRequestsPerMinute: 10,
		SensitiveBurstSize:         2,

		// 1-minute sliding window
		WindowDuration: time.Minute,
	}
}

// RateLimitError represents a rate limit violation
type RateLimitError struct {
	LimitType    LimitType
	Identifier   string
	Limit        int
	WindowStart  time.Time
	RetryAfter   time.Duration
	CurrentCount int
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded for %s '%s': %d/%d requests in window, retry after %v",
		e.LimitType, e.Identifier, e.CurrentCount, e.Limit, e.RetryAfter)
}

// Limiter defines the interface for rate limiting operations
type Limiter interface {
	// CheckUserLimit checks if a user has exceeded their rate limit
	CheckUserLimit(ctx context.Context, userID string, category EndpointCategory) error

	// CheckClientLimit checks if an OAuth client has exceeded their rate limit
	CheckClientLimit(ctx context.Context, clientID string) error

	// CheckIPLimit checks if an IP address has exceeded their rate limit for a category
	CheckIPLimit(ctx context.Context, ip string, category EndpointCategory) error

	// GetUserUsage returns current usage for a user
	GetUserUsage(ctx context.Context, userID string, category EndpointCategory) (current int, limit int, err error)

	// GetClientUsage returns current usage for a client
	GetClientUsage(ctx context.Context, clientID string) (current int, limit int, err error)

	// GetIPUsage returns current usage for an IP
	GetIPUsage(ctx context.Context, ip string, category EndpointCategory) (current int, limit int, err error)

	// ResetUserLimit manually resets a user's rate limit (admin function)
	ResetUserLimit(ctx context.Context, userID string, category EndpointCategory) error

	// ResetIPLimit manually resets an IP's rate limit (admin function)
	ResetIPLimit(ctx context.Context, ip string, category EndpointCategory) error
}
