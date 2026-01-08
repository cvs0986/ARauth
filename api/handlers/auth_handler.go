package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/auth/login"
	"github.com/arauth-identity/iam/auth/token"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	loginService   login.ServiceInterface
	refreshService *token.RefreshService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(loginService login.ServiceInterface, refreshService *token.RefreshService) *AuthHandler {
	return &AuthHandler{
		loginService:   loginService,
		refreshService: refreshService,
	}
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req login.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Try to get tenant ID from context (may not exist for SYSTEM users)
	// For SYSTEM users, tenant_id is not required
	tenantID, exists := middleware.GetTenantID(c)
	if exists {
		req.TenantID = tenantID
	}
	// If tenant_id doesn't exist in context, it will remain uuid.Nil
	// Login service will handle SYSTEM users (no tenant_id required)

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

// RefreshToken handles POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req token.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	resp, err := h.refreshService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "token_refresh_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// RevokeToken handles POST /api/v1/auth/revoke
func (h *AuthHandler) RevokeToken(c *gin.Context) {
	var req struct {
		Token      string `json:"token" binding:"required"`
		TokenType  string `json:"token_type_hint,omitempty"` // "access_token" or "refresh_token"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Determine token type
	if req.TokenType == "refresh_token" || req.TokenType == "" {
		// Try to revoke as refresh token
		if err := h.refreshService.RevokeRefreshToken(c.Request.Context(), req.Token); err == nil {
			c.JSON(http.StatusOK, gin.H{"message": "Token revoked successfully"})
			return
		}
	}

	// If refresh token revocation failed or it's an access token, add to blacklist
	// TODO: Implement Redis blacklist for access tokens
	c.JSON(http.StatusOK, gin.H{
		"message": "Token revocation requested. Access tokens will be invalidated on expiry.",
	})
}

