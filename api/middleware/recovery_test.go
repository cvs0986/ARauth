package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestRecovery_PanicHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zaptest.NewLogger(t)
	router := gin.New()
	router.Use(Recovery(logger))
	router.GET("/test", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Recovery middleware should catch panic and return 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal_server_error")
}

func TestRecovery_NormalRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zaptest.NewLogger(t)
	router := gin.New()
	router.Use(Recovery(logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Normal request should work fine
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRecovery_StringPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zaptest.NewLogger(t)
	router := gin.New()
	router.Use(Recovery(logger))
	router.GET("/test", func(c *gin.Context) {
		panic("string panic message")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRecovery_ErrorPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zaptest.NewLogger(t)
	router := gin.New()
	router.Use(Recovery(logger))
	router.GET("/test", func(c *gin.Context) {
		panic(assert.AnError)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

