package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuage-identity/iam/auth/login"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	loginService *login.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(loginService *login.Service) *AuthHandler {
	return &AuthHandler{loginService: loginService}
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req login.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

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

