package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arauth-identity/iam/auth/token"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthHandler_RevokeToken_Failures(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockTokenService := new(MockTokenService)
	mockAuditService := new(MockAuditService)

	// Create real RefreshService with mock dependencies
	refreshService := token.NewRefreshService(mockTokenService, nil, nil, nil, nil)

	handler := NewAuthHandler(nil, refreshService, mockTokenService, mockAuditService, nil)

	router := gin.New()
	router.POST("/api/v1/auth/revoke", handler.RevokeToken)

	// Scenario 1: No Token Provided
	t.Run("NoToken", func(t *testing.T) {
		// Handler tries to revoke as refresh token first (with empty string), so we must expect hashing failure
		mockTokenService.On("HashRefreshToken", mock.Anything).Return("", assert.AnError).Once()

		req, _ := http.NewRequest("POST", "/api/v1/auth/revoke", bytes.NewBufferString("{}"))
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		// Should contain error message
		assert.Contains(t, w.Body.String(), "Token not provided")
	})

	// Scenario 2: Revocation fails (redis error)
	t.Run("RevocationError", func(t *testing.T) {
		tokenVal := "error-token"

		// Falls through refresh check
		mockTokenService.On("HashRefreshToken", mock.Anything).Return("", assert.AnError).Once()

		// Access token revocation fails
		mockTokenService.On("RevokeAccessToken", mock.Anything, tokenVal).Return(assert.AnError).Once()

		// Audit log still attempts to validate token to get actor info
		mockTokenService.On("ValidateAccessToken", tokenVal).Return(nil, assert.AnError).Once()

		req, _ := http.NewRequest("POST", "/api/v1/auth/revoke", bytes.NewBufferString("{}"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenVal)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Current implementation returns 200 OK even on error (idempotent/safe logout)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
