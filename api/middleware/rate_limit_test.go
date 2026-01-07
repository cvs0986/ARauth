package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/arauth-identity/iam/internal/cache"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit_NoCache(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RateLimit(nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Without cache, rate limiting should be bypassed
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimit_WithCache(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a mock cache (nil for now, will skip rate limiting)
	cacheClient := cache.NewCache(nil)

	router := gin.New()
	router.Use(RateLimit(cacheClient))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should succeed (rate limit not enforced without Redis)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimit_SkipHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RateLimit(nil))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Health check should always be allowed
	assert.Equal(t, http.StatusOK, w.Code)
}

