package handlers

import (
	"net/http"

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
	// Get user ID from JWT token claims (set by JWTAuthMiddleware)
	claimsObj, exists := c.Get("user_claims")
	if !exists {
		middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
			"User claims not found", nil)
		c.Abort()
		return
	}

	userClaims := claimsObj.(*claims.Claims)
	userID, err := uuid.Parse(userClaims.Subject)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_user_id",
			"Invalid user ID in token", nil)
		return
	}

	// Create enroll request with user ID from token
	req := &mfa.EnrollRequest{
		UserID: userID,
	}

	resp, err := h.mfaService.Enroll(c.Request.Context(), req)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "enrollment_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Verify handles POST /api/v1/mfa/verify
func (h *MFAHandler) Verify(c *gin.Context) {
	// Get user ID from JWT token claims (set by JWTAuthMiddleware)
	claimsObj, exists := c.Get("user_claims")
	if !exists {
		middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
			"User claims not found", nil)
		c.Abort()
		return
	}

	userClaims := claimsObj.(*claims.Claims)
	userID, err := uuid.Parse(userClaims.Subject)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_user_id",
			"Invalid user ID in token", nil)
		return
	}

	// Parse request body for TOTP code
	var body struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Create verify request with user ID from token
	req := &mfa.VerifyRequest{
		UserID:   userID,
		TOTPCode: body.Code,
	}

	valid, err := h.mfaService.Verify(c.Request.Context(), req)
	if err != nil {
		// Log the actual error for debugging
		// Check if it's a user not found, MFA not enabled, or secret not found error
		if err.Error() == "MFA is not enabled for this user" {
			middleware.RespondWithError(c, http.StatusBadRequest, "mfa_not_enabled",
				"MFA is not enabled for this user. Please enroll in MFA first.", nil)
			return
		}
		if err.Error() == "MFA secret not found" {
			middleware.RespondWithError(c, http.StatusBadRequest, "mfa_secret_not_found",
				"MFA secret not found. Please re-enroll in MFA.", nil)
			return
		}
		middleware.RespondWithError(c, http.StatusUnauthorized, "verification_failed",
			err.Error(), nil)
		return
	}

	if !valid {
		middleware.RespondWithError(c, http.StatusUnauthorized, "invalid_code",
			"Invalid TOTP code or recovery code. Please check your authenticator app and try again.", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"verified": true,
	})
}

// Challenge handles POST /api/v1/mfa/challenge
func (h *MFAHandler) Challenge(c *gin.Context) {
	// Parse request body - tenant_id is optional for SYSTEM users
	var body struct {
		UserID   string `json:"user_id" binding:"required"`
		TenantID string `json:"tenant_id"` // Optional - empty for SYSTEM users
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	userID, err := uuid.Parse(body.UserID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_user_id",
			"Invalid user ID format", nil)
		return
	}

	// For SYSTEM users, tenant_id might be empty
	var tenantID uuid.UUID
	if body.TenantID != "" {
		tenantID, err = uuid.Parse(body.TenantID)
		if err != nil {
			middleware.RespondWithError(c, http.StatusBadRequest, "invalid_tenant_id",
				"Invalid tenant ID format", nil)
			return
		}
	}

	req := &mfa.ChallengeRequest{
		UserID:   userID,
		TenantID: tenantID,
	}

	resp, err := h.mfaService.CreateChallenge(c.Request.Context(), req)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "challenge_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// VerifyChallenge handles POST /api/v1/mfa/challenge/verify
func (h *MFAHandler) VerifyChallenge(c *gin.Context) {
	// Parse request body - support both challenge_id/code and session_id/totp_code formats
	var body struct {
		ChallengeID string `json:"challenge_id"` // Frontend uses this
		Code        string `json:"code"`          // Frontend uses this
		SessionID   string `json:"session_id"`   // Backend expects this
		TOTPCode    string `json:"totp_code"`     // Backend expects this
		RecoveryCode string `json:"recovery_code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Map frontend format (challenge_id/code) to backend format (session_id/totp_code)
	sessionID := body.SessionID
	if sessionID == "" {
		sessionID = body.ChallengeID // Use challenge_id if session_id not provided
	}
	if sessionID == "" {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"session_id or challenge_id is required", nil)
		return
	}

	totpCode := body.TOTPCode
	if totpCode == "" {
		totpCode = body.Code // Use code if totp_code not provided
	}

	// Create verify challenge request
	req := &mfa.VerifyChallengeRequest{
		SessionID:   sessionID,
		TOTPCode:    totpCode,
		RecoveryCode: body.RecoveryCode,
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

	// Note: Refresh token storage is skipped in MFA flow
	// The refresh token is returned to the client but not stored in the database
	// This can be enhanced later if refresh token storage is needed for MFA flow

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

