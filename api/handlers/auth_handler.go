package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuage-identity/iam/api/middleware"
	"github.com/nuage-identity/iam/auth/login"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	loginService login.ServiceInterface
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(loginService login.ServiceInterface) *AuthHandler {
	return &AuthHandler{loginService: loginService}
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	// Get tenant ID from context (set by tenant middleware)
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	var req login.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Set tenant ID from context (always override any tenant_id in request body for security)
	req.TenantID = tenantID

	resp, err := h.loginService.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "authentication_failed",
			"message": err.Error(),
		})
		return
	}

	if resp.MFARequired {
		c.JSON(http.StatusOK, resp)
		return
	}

	if resp.RedirectTo != "" {
		c.JSON(http.StatusOK, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
}

