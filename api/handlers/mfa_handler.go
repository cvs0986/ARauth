package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuage-identity/iam/api/middleware"
	"github.com/nuage-identity/iam/auth/mfa"
)

// MFAHandler handles MFA-related HTTP requests
type MFAHandler struct {
	mfaService *mfa.Service
}

// NewMFAHandler creates a new MFA handler
func NewMFAHandler(mfaService *mfa.Service) *MFAHandler {
	return &MFAHandler{mfaService: mfaService}
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
		middleware.RespondWithError(c, http.StatusUnauthorized, "verification_failed",
			err.Error(), nil)
		return
	}

	if !resp.Verified {
		middleware.RespondWithError(c, http.StatusUnauthorized, "invalid_code",
			"Invalid TOTP code or recovery code", nil)
		return
	}

	c.JSON(http.StatusOK, resp)
}

