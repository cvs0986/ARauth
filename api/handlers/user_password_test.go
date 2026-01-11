package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserHandler_ChangePassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockUserService)
	mockAuditService := new(MockAuditService)
	// Passing nil for SystemRoleRepo and RoleRepo as we test Tenant User flow where they are not used immediately
	// (Except SystemRoleRepo is used inside the handler logic ONLY if PrincipalType == System)
	handler := NewUserHandler(mockService, nil, nil, mockAuditService)

	tenantID := uuid.New()
	userID := uuid.New()
	targetUserID := uuid.New()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate auth middleware and tenant middleware
		c.Set("user_claims", &claims.Claims{Subject: userID.String()})
		c.Set("tenant_id", tenantID)
		c.Next()
	})
	router.POST("/users/:id/change-password", handler.ChangePassword)

	t.Run("success", func(t *testing.T) {
		targetUser := &models.User{
			ID:            targetUserID,
			TenantID:      &tenantID,
			PrincipalType: models.PrincipalTypeTenant,
			Username:      "targetuser",
		}

		// Reset mocks because we are reusing them or sharing instance?
		// Better to re-instantiate or just append expectations carefully.
		// Since we reuse, let's keep it simple for this one run.

		mockService.On("GetByID", mock.Anything, targetUserID).Return(targetUser, nil)
		mockService.On("ChangePassword", mock.Anything, targetUserID, "NewPassword123!").Return(nil)

		// Expect audit log
		mockAuditService.On("LogUserUpdated", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		reqBody := map[string]string{"password": "NewPassword123!"}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/users/"+targetUserID.String()+"/change-password", bytes.NewBuffer(jsonBody))
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		mockService.AssertExpectations(t)
	})
}
