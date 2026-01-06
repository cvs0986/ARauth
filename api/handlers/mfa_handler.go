package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nuage-identity/iam/api/middleware"
	"github.com/nuage-identity/iam/auth/mfa"
	"github.com/nuage-identity/iam/internal/audit"
)

// MFAHandler handles MFA-related HTTP requests
type MFAHandler struct {
	mfaService  *mfa.Service
	auditLogger *audit.Logger
}

// NewMFAHandler creates a new MFA handler
func NewMFAHandler(mfaService *mfa.Service, auditLogger *audit.Logger) *MFAHandler {
	return &MFAHandler{
		mfaService:  mfaService,
		auditLogger: auditLogger,
	}
}

// Enroll handles POST /api/v1/mfa/enroll
func (h *MFAHandler) Enroll(c *gin.Context) {
	var req mfa.EnrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	resp, err := h.mfaService.Enroll(c.Request.Context(), &req)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "enrollment_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Verify handles POST /api/v1/mfa/verify
func (h *MFAHandler) Verify(c *gin.Context) {
	var req mfa.VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	valid, err := h.mfaService.Verify(c.Request.Context(), &req)
	if err != nil {
		middleware.RespondWithError(c, http.StatusUnauthorized, "verification_failed",
			err.Error(), nil)
		return
	}

	if !valid {
		middleware.RespondWithError(c, http.StatusUnauthorized, "invalid_code",
			"Invalid TOTP code or recovery code", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"verified": true,
	})
}

// Challenge handles POST /api/v1/mfa/challenge
func (h *MFAHandler) Challenge(c *gin.Context) {
	var req mfa.ChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	resp, err := h.mfaService.CreateChallenge(c.Request.Context(), &req)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "challenge_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// VerifyChallenge handles POST /api/v1/mfa/challenge/verify
func (h *MFAHandler) VerifyChallenge(c *gin.Context) {
	var req mfa.VerifyChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	resp, err := h.mfaService.VerifyChallenge(c.Request.Context(), &req)
	if err != nil {
		// Log failed attempt
		if resp != nil && resp.UserID != "" {
			// Try to log if we have user info
		}
		middleware.RespondWithError(c, http.StatusUnauthorized, "verification_failed",
			err.Error(), nil)
		return
	}

	if !resp.Verified {
		// Log failed verification
		if resp.UserID != "" {
			userID, _ := uuid.Parse(resp.UserID)
			tenantID, _ := uuid.Parse(resp.TenantID)
			h.auditLogger.LogMFAAction(c.Request.Context(), tenantID, userID, "verify_challenge", c.Request, "failure", "Invalid MFA code")
		}
		middleware.RespondWithError(c, http.StatusUnauthorized, "invalid_code",
			"Invalid TOTP code or recovery code", nil)
		return
	}

	// Log successful verification
	if resp.UserID != "" {
		userID, _ := uuid.Parse(resp.UserID)
		tenantID, _ := uuid.Parse(resp.TenantID)
		h.auditLogger.LogMFAAction(c.Request.Context(), tenantID, userID, "verify_challenge", c.Request, "success", "MFA challenge verified")
	}

	c.JSON(http.StatusOK, resp)
}

