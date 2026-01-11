package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/arauth-identity/iam/identity/ratelimit"
	"github.com/gin-gonic/gin"
)

// MultiTierRateLimit creates a comprehensive rate limiting middleware
// that applies user, client, and IP-based limits based on context
func MultiTierRateLimit(limiter ratelimit.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip health checks
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/health/live" || c.Request.URL.Path == "/health/ready" {
			c.Next()
			return
		}

		// Determine endpoint category
		category := categorizeEndpoint(c.Request.URL.Path)

		// Check IP-based rate limit first (prevents DDoS)
		ip := c.ClientIP()
		if err := limiter.CheckIPLimit(c.Request.Context(), ip, category); err != nil {
			handleRateLimitError(c, err)
			return
		}

		// Check user-based rate limit if user is authenticated
		if userID, exists := c.Get("user_id"); exists {
			if err := limiter.CheckUserLimit(c.Request.Context(), fmt.Sprintf("%v", userID), category); err != nil {
				handleRateLimitError(c, err)
				return
			}
		}

		// Check client-based rate limit if OAuth client is present
		if clientID, exists := c.Get("client_id"); exists {
			if err := limiter.CheckClientLimit(c.Request.Context(), fmt.Sprintf("%v", clientID)); err != nil {
				handleRateLimitError(c, err)
				return
			}
		}

		c.Next()
	}
}

// UserOnlyRateLimit applies rate limiting only to authenticated users
func UserOnlyRateLimit(limiter ratelimit.Limiter, category ratelimit.EndpointCategory) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			// No user context, skip rate limiting
			c.Next()
			return
		}

		if err := limiter.CheckUserLimit(c.Request.Context(), fmt.Sprintf("%v", userID), category); err != nil {
			handleRateLimitError(c, err)
			return
		}

		c.Next()
	}
}

// IPOnlyRateLimit applies rate limiting based on IP address only
func IPOnlyRateLimit(limiter ratelimit.Limiter, category ratelimit.EndpointCategory) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if err := limiter.CheckIPLimit(c.Request.Context(), ip, category); err != nil {
			handleRateLimitError(c, err)
			return
		}

		c.Next()
	}
}

// ClientOnlyRateLimit applies rate limiting based on OAuth client only
func ClientOnlyRateLimit(limiter ratelimit.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID, exists := c.Get("client_id")
		if !exists {
			// No client context, skip rate limiting
			c.Next()
			return
		}

		if err := limiter.CheckClientLimit(c.Request.Context(), fmt.Sprintf("%v", clientID)); err != nil {
			handleRateLimitError(c, err)
			return
		}

		c.Next()
	}
}

// categorizeEndpoint determines the rate limit category based on the endpoint path
func categorizeEndpoint(path string) ratelimit.EndpointCategory {
	// Auth endpoints (login, token, etc.)
	if matchesPrefix(path, []string{"/api/v1/auth/login", "/api/v1/auth/token", "/api/v1/auth/refresh"}) {
		return ratelimit.CategoryAuth
	}

	// Sensitive endpoints (MFA, password reset, etc.)
	if matchesPrefix(path, []string{"/api/v1/auth/mfa"}) ||
		strings.Contains(path, "/reset-password") ||
		strings.Contains(path, "/reset-mfa") ||
		strings.Contains(path, "/suspend") {
		return ratelimit.CategorySensitive
	}

	// Admin endpoints
	if matchesPrefix(path, []string{
		"/api/v1/tenants",
		"/api/v1/system",
		"/api/v1/audit",
		"/api/v1/impersonation",
	}) {
		return ratelimit.CategoryAdmin
	}

	// Default to general category
	return ratelimit.CategoryGeneral
}

// matchesPrefix checks if path matches any of the given prefixes
func matchesPrefix(path string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// handleRateLimitError handles rate limit errors and sends appropriate response
func handleRateLimitError(c *gin.Context, err error) {
	rateLimitErr, ok := err.(*ratelimit.RateLimitError)
	if !ok {
		// Generic error
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":   "rate_limit_exceeded",
			"message": "Too many requests. Please try again later.",
		})
		c.Abort()
		return
	}

	// Set rate limit headers
	c.Header("X-RateLimit-Limit", strconv.Itoa(rateLimitErr.Limit))
	c.Header("X-RateLimit-Remaining", "0")
	c.Header("X-RateLimit-Reset", strconv.FormatInt(rateLimitErr.WindowStart.Add(rateLimitErr.RetryAfter).Unix(), 10))
	c.Header("Retry-After", strconv.FormatInt(int64(rateLimitErr.RetryAfter.Seconds()), 10))

	c.JSON(http.StatusTooManyRequests, gin.H{
		"error":   "rate_limit_exceeded",
		"message": fmt.Sprintf("Rate limit exceeded. Please retry after %v.", rateLimitErr.RetryAfter.Round(time.Second)),
		"details": gin.H{
			"limit_type":    string(rateLimitErr.LimitType),
			"limit":         rateLimitErr.Limit,
			"current_count": rateLimitErr.CurrentCount,
			"retry_after":   rateLimitErr.RetryAfter.Seconds(),
		},
	})
	c.Abort()
}
