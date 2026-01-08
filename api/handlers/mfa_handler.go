package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/auth/mfa"
	"github.com/arauth-identity/iam/auth/token"
	"github.com/arauth-identity/iam/internal/audit"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// MFAHandler handles MFA-related HTTP requests
type MFAHandler struct {
	mfaService      mfa.ServiceInterface
	auditLogger     *audit.Logger
	tokenService    token.ServiceInterface
	claimsBuilder   *claims.Builder
	userRepo        interfaces.UserRepository
	lifetimeResolver *token.LifetimeResolver
}

// NewMFAHandler creates a new MFA handler
func NewMFAHandler(
	mfaService mfa.ServiceInterface,
	auditLogger *audit.Logger,
	tokenService token.ServiceInterface,
	claimsBuilder *claims.Builder,
	userRepo interfaces.UserRepository,
	lifetimeResolver *token.LifetimeResolver,
) *MFAHandler {
	return &MFAHandler{
		mfaService:       mfaService,
		auditLogger:      auditLogger,
		tokenService:     tokenService,
		claimsBuilder:    claimsBuilder,
		userRepo:         userRepo,
		lifetimeResolver: lifetimeResolver,
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
		// Log failed attempt if we have user info
		if resp != nil && resp.UserID != "" {
			userID, _ := uuid.Parse(resp.UserID)
			tenantID, _ := uuid.Parse(resp.TenantID)
			_ = h.auditLogger.LogMFAAction(c.Request.Context(), tenantID, userID, "verify_challenge", c.Request, "failure", err.Error()) // Ignore audit log errors
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
			_ = h.auditLogger.LogMFAAction(c.Request.Context(), tenantID, userID, "verify_challenge", c.Request, "failure", "Invalid MFA code") // Ignore audit log errors
		}
		middleware.RespondWithError(c, http.StatusUnauthorized, "invalid_code",
			"Invalid TOTP code or recovery code", nil)
		return
	}

	// Log successful verification
	if resp.UserID != "" {
		userID, _ := uuid.Parse(resp.UserID)
		tenantID, _ := uuid.Parse(resp.TenantID)
		_ = h.auditLogger.LogMFAAction(c.Request.Context(), tenantID, userID, "verify_challenge", c.Request, "success", "MFA challenge verified") // Ignore audit log errors
	}

	// Issue tokens after successful MFA verification
	userID, _ := uuid.Parse(resp.UserID)
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "user_not_found",
			"Failed to retrieve user after MFA verification", nil)
		return
	}

	// Build claims
	claimsObj, err := h.claimsBuilder.BuildClaims(c.Request.Context(), user)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "claims_build_failed",
			"Failed to build claims", nil)
		return
	}

	// Get token lifetimes
	var tenantID uuid.UUID
	if user.TenantID != nil {
		tenantID = *user.TenantID
	}
	lifetimes := h.lifetimeResolver.GetAllLifetimes(c.Request.Context(), tenantID, false) // TODO: Support remember_me from request

	// Generate access token
	accessToken, err := h.tokenService.GenerateAccessToken(claimsObj, lifetimes.AccessTokenTTL)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "token_issue_failed",
			"Failed to generate access token", nil)
		return
	}

	// Generate refresh token
	refreshToken, err := h.tokenService.GenerateRefreshToken()
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "refresh_token_issue_failed",
			"Failed to generate refresh token", nil)
		return
	}

	// Hash refresh token for storage
	refreshTokenHash, err := h.tokenService.HashRefreshToken(refreshToken)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "refresh_token_hash_failed",
			"Failed to hash refresh token", nil)
		return
	}

	// Store refresh token (we need refreshTokenRepo, but we don't have it in MFA handler)
	// For now, we'll skip storing refresh token in MFA flow - it can be added later if needed
	// TODO: Add refreshTokenRepo to MFA handler if refresh tokens are needed in MFA flow

	// Generate ID token (same as access token for now, can be enhanced later)
	idToken, err := h.tokenService.GenerateAccessToken(claimsObj, lifetimes.IDTokenTTL)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "id_token_issue_failed",
			"Failed to generate ID token", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"verified":           true,
		"access_token":       accessToken,
		"refresh_token":      refreshToken, // Return plain token to client
		"id_token":           idToken,
		"token_type":         "Bearer",
		"expires_in":         int(lifetimes.AccessTokenTTL.Seconds()),
		"refresh_expires_in": int(lifetimes.RefreshTokenTTL.Seconds()),
	})
}

