package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_Check(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &HealthHandler{}
	router.GET("/health", handler.Check)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthHandler_Live(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &HealthHandler{}
	router.GET("/health/live", handler.Live)

	req, _ := http.NewRequest("GET", "/health/live", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthHandler_Ready(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &HealthHandler{}
	router.GET("/health/ready", handler.Ready)

	req, _ := http.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

