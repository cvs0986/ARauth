package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// BenchmarkHealthCheck benchmarks the health check endpoint
func BenchmarkHealthCheck(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	handler := NewHealthHandler()

	router := gin.New()
	router.GET("/health", handler.Check)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)
	}
}

// BenchmarkHealthLive benchmarks the liveness endpoint
func BenchmarkHealthLive(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	handler := NewHealthHandler()

	router := gin.New()
	router.GET("/health/live", handler.Liveness)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health/live", nil)
		router.ServeHTTP(w, req)
	}
}

// BenchmarkHealthReady benchmarks the readiness endpoint
func BenchmarkHealthReady(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	handler := NewHealthHandler()

	router := gin.New()
	router.GET("/health/ready", handler.Readiness)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health/ready", nil)
		router.ServeHTTP(w, req)
	}
}

