package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/identity/linking"
)

// IdentityLinkingHandler handles identity linking HTTP requests
type IdentityLinkingHandler struct {
	linkingService linking.ServiceInterface
}

// NewIdentityLinkingHandler creates a new identity linking handler
func NewIdentityLinkingHandler(linkingService linking.ServiceInterface) *IdentityLinkingHandler {
	return &IdentityLinkingHandler{
		linkingService: linkingService,
	}
}

// LinkIdentity handles POST /api/v1/users/:id/identities
func (h *IdentityLinkingHandler) LinkIdentity(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid user ID"})
		return
	}

	var req struct {
		ProviderID uuid.UUID              `json:"provider_id" binding:"required"`
		ExternalID string                 `json:"external_id" binding:"required"`
		Attributes map[string]interface{} `json:"attributes,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	if err := h.linkingService.LinkIdentity(c.Request.Context(), userID, req.ProviderID, req.ExternalID, req.Attributes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "link_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Identity linked successfully"})
}

// UnlinkIdentity handles DELETE /api/v1/users/:id/identities/:identity_id
func (h *IdentityLinkingHandler) UnlinkIdentity(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid user ID"})
		return
	}

	identityIDStr := c.Param("identity_id")
	identityID, err := uuid.Parse(identityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid identity ID"})
		return
	}

	if err := h.linkingService.UnlinkIdentity(c.Request.Context(), userID, identityID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unlink_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// SetPrimaryIdentity handles PUT /api/v1/users/:id/identities/:identity_id/primary
func (h *IdentityLinkingHandler) SetPrimaryIdentity(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid user ID"})
		return
	}

	identityIDStr := c.Param("identity_id")
	identityID, err := uuid.Parse(identityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid identity ID"})
		return
	}

	if err := h.linkingService.SetPrimaryIdentity(c.Request.Context(), userID, identityID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "set_primary_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Primary identity set successfully"})
}

// GetUserIdentities handles GET /api/v1/users/:id/identities
func (h *IdentityLinkingHandler) GetUserIdentities(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid user ID"})
		return
	}

	identities, err := h.linkingService.GetUserIdentities(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, identities)
}

// VerifyIdentity handles POST /api/v1/users/:id/identities/:identity_id/verify
func (h *IdentityLinkingHandler) VerifyIdentity(c *gin.Context) {
	identityIDStr := c.Param("identity_id")
	identityID, err := uuid.Parse(identityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid identity ID"})
		return
	}

	if err := h.linkingService.VerifyIdentity(c.Request.Context(), identityID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "verify_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Identity verified successfully"})
}

