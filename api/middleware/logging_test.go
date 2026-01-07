package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestLogging(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zaptest.NewLogger(t)
	router := gin.New()
	router.Use(Logging(logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Logging middleware should not affect response
}

func TestLogging_WithQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zaptest.NewLogger(t)
	router := gin.New()
	router.Use(Logging(logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?param=value", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogging_ErrorStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zaptest.NewLogger(t)
	router := gin.New()
	router.Use(Logging(logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

