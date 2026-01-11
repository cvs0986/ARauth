package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/arauth-identity/iam/identity/ratelimit"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestLimiter(t *testing.T) (ratelimit.Limiter, func()) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	config := &ratelimit.Config{
		UserRequestsPerMinute:      5,
		UserBurstSize:              2,
		ClientRequestsPerMinute:    10,
		ClientBurstSize:            3,
		AdminIPRequestsPerMinute:   3,
		AdminIPBurstSize:           1,
		AuthRequestsPerMinute:      2,
		AuthBurstSize:              1,
		SensitiveRequestsPerMinute: 1,
		SensitiveBurstSize:         0,
		WindowDuration:             time.Minute,
	}

	limiter := ratelimit.NewRedisLimiter(client, config)

	cleanup := func() {
		client.Close()
		mr.Close()
	}

	return limiter, cleanup
}

func TestMultiTierRateLimit_IPLimit(t *testing.T) {
	limiter, cleanup := setupTestLimiter(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(MultiTierRateLimit(limiter))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make requests from same IP
	for i := 0; i < 7; i++ { // 5 + 2 burst
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "request %d should succeed", i+1)
	}

	// 8th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Header().Get("Retry-After"), "")
}

func TestMultiTierRateLimit_UserLimit(t *testing.T) {
	limiter, cleanup := setupTestLimiter(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(MultiTierRateLimit(limiter))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make requests as authenticated user
	for i := 0; i < 7; i++ { // 5 + 2 burst
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", "user-123")

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "request %d should succeed", i+1)
	}

	// 8th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", "user-123")

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestMultiTierRateLimit_CategoryAuth(t *testing.T) {
	limiter, cleanup := setupTestLimiter(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set user context before rate limiting
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-auth-test")
		c.Next()
	})
	router.Use(UserOnlyRateLimit(limiter, ratelimit.CategoryAuth))

	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Auth category has stricter limits (2 + 1 burst = 3) for users
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "request %d should succeed", i+1)
	}

	// 4th request should be rate limited (user limit exceeded)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestMultiTierRateLimit_SkipHealthCheck(t *testing.T) {
	limiter, cleanup := setupTestLimiter(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(MultiTierRateLimit(limiter))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Health checks should never be rate limited
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestUserOnlyRateLimit(t *testing.T) {
	limiter, cleanup := setupTestLimiter(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(UserOnlyRateLimit(limiter, ratelimit.CategoryGeneral))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Unauthenticated requests should not be rate limited
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestIPOnlyRateLimit(t *testing.T) {
	limiter, cleanup := setupTestLimiter(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(IPOnlyRateLimit(limiter, ratelimit.CategoryAdmin))
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Admin category has strict IP limits (3 + 1 burst = 4)
	for i := 0; i < 4; i++ {
		req := httptest.NewRequest("GET", "/admin", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "request %d should succeed", i+1)
	}

	// 5th request should be rate limited
	req := httptest.NewRequest("GET", "/admin", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestCategorizeEndpoint(t *testing.T) {
	tests := []struct {
		path     string
		expected ratelimit.EndpointCategory
	}{
		{"/api/v1/auth/login", ratelimit.CategoryAuth},
		{"/api/v1/auth/token", ratelimit.CategoryAuth},
		{"/api/v1/auth/mfa/enroll", ratelimit.CategorySensitive},
		{"/api/v1/users/123/reset-password", ratelimit.CategorySensitive},
		{"/api/v1/tenants", ratelimit.CategoryAdmin},
		{"/api/v1/audit/logs", ratelimit.CategoryAdmin},
		{"/api/v1/users", ratelimit.CategoryGeneral},
		{"/api/v1/roles", ratelimit.CategoryGeneral},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			category := categorizeEndpoint(tt.path)
			assert.Equal(t, tt.expected, category)
		})
	}
}
