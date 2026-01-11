package handlers

import (
	"net/http"

	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/identity/audit"
	"github.com/arauth-identity/iam/identity/scim"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SCIMTokenHandler handles SCIM token management HTTP requests
type SCIMTokenHandler struct {
	tokenService scim.TokenServiceInterface
	auditService audit.ServiceInterface
}

// NewSCIMTokenHandler creates a new SCIM token handler
func NewSCIMTokenHandler(tokenService scim.TokenServiceInterface, auditService audit.ServiceInterface) *SCIMTokenHandler {
	return &SCIMTokenHandler{
		tokenService: tokenService,
		auditService: auditService,
	}
}

// CreateToken handles POST /api/v1/scim/tokens
func (h *SCIMTokenHandler) CreateToken(c *gin.Context) {
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	var req scim.CreateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Invalid request body", nil)
		return
	}

	// Create token
	token, plaintext, err := h.tokenService.CreateToken(c.Request.Context(), tenantID, &req)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "creation_failed",
			"Failed to create SCIM token", nil)
		return
	}

	// Audit log
	if actor, err := extractActorFromContext(c); err == nil {
		sourceIP, userAgent := extractSourceInfo(c)
		_ = h.auditService.LogTokenIssued(c.Request.Context(), actor, &tenantID, sourceIP, userAgent, map[string]interface{}{
			"token_id": token.ID.String(),
			"name":     token.Name,
			"scopes":   token.Scopes,
		})
	}

	// Return response with plaintext token
	c.JSON(http.StatusCreated, gin.H{
		"token":           token, // Contains metadata
		"plaintext_token": plaintext,
	})
}

// ListTokens handles GET /api/v1/scim/tokens
func (h *SCIMTokenHandler) ListTokens(c *gin.Context) {
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	tokens, err := h.tokenService.ListTokens(c.Request.Context(), tenantID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "list_failed",
			"Failed to list SCIM tokens", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tokens": tokens,
		"count":  len(tokens),
	})
}

// GetToken handles GET /api/v1/scim/tokens/:id
func (h *SCIMTokenHandler) GetToken(c *gin.Context) {
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	tokenID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid token ID format", nil)
		return
	}

	token, err := h.tokenService.GetToken(c.Request.Context(), tokenID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"SCIM token not found", nil)
		return
	}

	// Verify ownership
	if token.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusForbidden, "forbidden",
			"Access denied to this SCIM token", nil)
		return
	}

	c.JSON(http.StatusOK, token)
}

// RotateToken handles POST /api/v1/scim/tokens/:id/rotate
func (h *SCIMTokenHandler) RotateToken(c *gin.Context) {
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	tokenID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid token ID format", nil)
		return
	}

	// Rotate token
	token, plaintext, err := h.tokenService.RotateToken(c.Request.Context(), tokenID)
	if err != nil {
		if err.Error() == "token not found" {
			middleware.RespondWithError(c, http.StatusNotFound, "not_found",
				"SCIM token not found", nil)
			return
		}
		middleware.RespondWithError(c, http.StatusInternalServerError, "rotation_failed",
			"Failed to rotate SCIM token", nil)
		return
	}

	// Verify ownership (if service doesn't enforce, handler must check retrieved token)
	if token.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusForbidden, "forbidden",
			"Access denied to this SCIM token", nil)
		return
	}

	// Audit log
	if actor, err := extractActorFromContext(c); err == nil {
		sourceIP, userAgent := extractSourceInfo(c)

		// Log new token issued for rotation
		_ = h.auditService.LogTokenIssued(c.Request.Context(), actor, &tenantID, sourceIP, userAgent, map[string]interface{}{
			"action":   "rotation",
			"token_id": token.ID.String(),
			"name":     token.Name,
		})
	}

	// Return response with new plaintext token
	c.JSON(http.StatusOK, gin.H{
		"token":           token,
		"plaintext_token": plaintext,
	})
}

// DeleteToken handles DELETE /api/v1/scim/tokens/:id
func (h *SCIMTokenHandler) DeleteToken(c *gin.Context) {
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	tokenID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid token ID format", nil)
		return
	}

	// Check existence and ownership first (optional optimization but safer)
	token, err := h.tokenService.GetToken(c.Request.Context(), tokenID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"SCIM token not found", nil)
		return
	}

	if token.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusForbidden, "forbidden",
			"Access denied to this SCIM token", nil)
		return
	}

	if err := h.tokenService.DeleteToken(c.Request.Context(), tokenID); err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "deletion_failed",
			"Failed to delete SCIM token", nil)
		return
	}

	// Audit log
	if actor, err := extractActorFromContext(c); err == nil {
		sourceIP, userAgent := extractSourceInfo(c)
		_ = h.auditService.LogTokenRevoked(c.Request.Context(), actor, &tenantID, sourceIP, userAgent, map[string]interface{}{
			"token_id": tokenID.String(),
			"name":     token.Name,
		})
	}

	c.Status(http.StatusNoContent)
}
