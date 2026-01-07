package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nuage-identity/iam/internal/cache"
)

// RateLimitConfig configures rate limiting behavior
type RateLimitConfig struct {
	// Requests per window
	Requests int
	// Time window duration
	Window time.Duration
	// Key generator function (defaults to IP address)
	KeyFunc func(*gin.Context) string
	// Skip function to bypass rate limiting
	SkipFunc func(*gin.Context) bool
	// Error handler
	ErrorHandler func(*gin.Context, error)
}

// DefaultRateLimitConfig returns default rate limit configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Requests: 100,
		Window:   time.Minute,
		KeyFunc: func(c *gin.Context) string {
			// Use IP address as default key
			ip := c.ClientIP()
			return fmt.Sprintf("rate_limit:%s", ip)
		},
		SkipFunc: func(c *gin.Context) bool {
			// Skip rate limiting for health checks
			return c.Request.URL.Path == "/health"
		},
		ErrorHandler: func(c *gin.Context, err error) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
			})
		},
	}
}

// RateLimit creates a rate limiting middleware
func RateLimit(cacheClient *cache.Cache) gin.HandlerFunc {
	config := DefaultRateLimitConfig()
	return RateLimitWithConfig(cacheClient, config)
}

// RateLimitWithConfig creates a rate limiting middleware with custom configuration
func RateLimitWithConfig(cacheClient *cache.Cache, config RateLimitConfig) gin.HandlerFunc {
	if cacheClient == nil {
		// If no cache client, return a no-op middleware
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Skip if skip function returns true
		if config.SkipFunc != nil && config.SkipFunc(c) {
			c.Next()
			return
		}

		// Generate rate limit key
		key := config.KeyFunc(c)
		windowKey := fmt.Sprintf("%s:window", key)
		countKey := fmt.Sprintf("%s:count", key)

		// Get current window timestamp
		now := time.Now()
		windowStart := now.Truncate(config.Window).Unix()

		// Try to get current window
		var currentWindow int64
		err := cacheClient.Get(c.Request.Context(), windowKey, &currentWindow)
		if err != nil {
			// Window doesn't exist, create new one
			currentWindow = windowStart
			_ = cacheClient.Set(c.Request.Context(), windowKey, windowStart, config.Window*2) // Ignore cache errors
			_ = cacheClient.Set(c.Request.Context(), countKey, 1, config.Window*2)             // Ignore cache errors
			c.Next()
			return
		}

		// Check if we're in a new window
		if currentWindow != windowStart {
			// New window, reset count
			currentWindow = windowStart
			_ = cacheClient.Set(c.Request.Context(), windowKey, windowStart, config.Window*2) // Ignore cache errors
			_ = cacheClient.Set(c.Request.Context(), countKey, 1, config.Window*2)             // Ignore cache errors
			c.Next()
			return
		}

		// Get current count
		var count int
		err = cacheClient.Get(c.Request.Context(), countKey, &count)
		if err != nil {
			// Count doesn't exist, start fresh
			_ = cacheClient.Set(c.Request.Context(), countKey, 1, config.Window*2) // Ignore cache errors
			c.Next()
			return
		}

		// Check if limit exceeded
		if count >= config.Requests {
			// Calculate retry after
			retryAfter := config.Window - time.Since(time.Unix(windowStart, 0))
			if retryAfter < 0 {
				retryAfter = 0
			}

			c.Header("X-RateLimit-Limit", strconv.Itoa(config.Requests))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(windowStart+int64(config.Window.Seconds()), 10))
			c.Header("Retry-After", strconv.FormatInt(int64(retryAfter.Seconds()), 10))

			if config.ErrorHandler != nil {
				config.ErrorHandler(c, fmt.Errorf("rate limit exceeded"))
			} else {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":   "rate_limit_exceeded",
					"message": "Too many requests. Please try again later.",
				})
			}
			c.Abort()
			return
		}

		// Increment count
		newCount := count + 1
		_ = cacheClient.Set(c.Request.Context(), countKey, newCount, config.Window*2) // Ignore cache errors

		// Set rate limit headers
		remaining := config.Requests - newCount
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(windowStart+int64(config.Window.Seconds()), 10))

		c.Next()
	}
}

// TenantRateLimit creates a rate limiting middleware scoped to tenant
func TenantRateLimit(cacheClient *cache.Cache, requestsPerMinute int) gin.HandlerFunc {
	config := DefaultRateLimitConfig()
	config.Requests = requestsPerMinute
	config.Window = time.Minute
	config.KeyFunc = func(c *gin.Context) string {
		// Get tenant ID from context
		tenantID, exists := GetTenantID(c)
		if !exists {
			// Fallback to IP if no tenant
			return fmt.Sprintf("rate_limit:tenant:unknown:%s", c.ClientIP())
		}
		return fmt.Sprintf("rate_limit:tenant:%s", tenantID.String())
	}

	return RateLimitWithConfig(cacheClient, config)
}

// UserRateLimit creates a rate limiting middleware scoped to user
func UserRateLimit(cacheClient *cache.Cache, requestsPerMinute int) gin.HandlerFunc {
	config := DefaultRateLimitConfig()
	config.Requests = requestsPerMinute
	config.Window = time.Minute
	config.KeyFunc = func(c *gin.Context) string {
		// Get user ID from context (set by auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			// Fallback to IP if no user
			return fmt.Sprintf("rate_limit:user:unknown:%s", c.ClientIP())
		}
		return fmt.Sprintf("rate_limit:user:%v", userID)
	}

	return RateLimitWithConfig(cacheClient, config)
}
