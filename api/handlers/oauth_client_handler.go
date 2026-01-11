package handlers

import (
	"net/http"

	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/identity/oauthclient"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OAuthClientHandler handles OAuth2 client management HTTP requests
type OAuthClientHandler struct {
	clientService oauthclient.ServiceInterface
}

// NewOAuthClientHandler creates a new OAuth client handler
func NewOAuthClientHandler(clientService oauthclient.ServiceInterface) *OAuthClientHandler {
	return &OAuthClientHandler{
		clientService: clientService,
	}
}

// CreateClient handles POST /api/v1/oauth/clients
func (h *OAuthClientHandler) CreateClient(c *gin.Context) {
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	var req oauthclient.CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Invalid request body", nil)
		return
	}

	// Get user ID for created_by
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
			"User ID not found in token", nil)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_user_id",
			"Invalid user ID format", nil)
		return
	}

	// Create client (service handles secret generation and hashing)
	resp, err := h.clientService.CreateClient(c.Request.Context(), tenantID, &req, userUUID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "creation_failed",
			"Failed to create OAuth client", nil)
		return
	}

	// Return response with ONE-TIME secret
	// SECURITY: This is the ONLY time the plaintext secret is returned
	c.JSON(http.StatusCreated, resp)
}

// ListClients handles GET /api/v1/oauth/clients
func (h *OAuthClientHandler) ListClients(c *gin.Context) {
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	clients, err := h.clientService.ListClients(c.Request.Context(), tenantID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "list_failed",
			"Failed to list OAuth clients", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"clients": clients,
		"count":   len(clients),
	})
}

// GetClient handles GET /api/v1/oauth/clients/:id
func (h *OAuthClientHandler) GetClient(c *gin.Context) {
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid client ID format", nil)
		return
	}

	client, err := h.clientService.GetClient(c.Request.Context(), clientID, tenantID)
	if err != nil {
		if err.Error() == "oauth client does not belong to tenant" {
			middleware.RespondWithError(c, http.StatusForbidden, "forbidden",
				"Access denied to this OAuth client", nil)
			return
		}
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"OAuth client not found", nil)
		return
	}

	c.JSON(http.StatusOK, client)
}

// RotateSecret handles POST /api/v1/oauth/clients/:id/rotate-secret
func (h *OAuthClientHandler) RotateSecret(c *gin.Context) {
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid client ID format", nil)
		return
	}

	// Rotate secret (service handles generation and hashing)
	resp, err := h.clientService.RotateSecret(c.Request.Context(), clientID, tenantID)
	if err != nil {
		if err.Error() == "oauth client does not belong to tenant" {
			middleware.RespondWithError(c, http.StatusForbidden, "forbidden",
				"Access denied to this OAuth client", nil)
			return
		}
		middleware.RespondWithError(c, http.StatusInternalServerError, "rotation_failed",
			"Failed to rotate client secret", nil)
		return
	}

	// Return response with ONE-TIME new secret
	// SECURITY: This is the ONLY time the new plaintext secret is returned
	c.JSON(http.StatusOK, resp)
}

// DeleteClient handles DELETE /api/v1/oauth/clients/:id
func (h *OAuthClientHandler) DeleteClient(c *gin.Context) {
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid client ID format", nil)
		return
	}

	// Delete client
	if err := h.clientService.DeleteClient(c.Request.Context(), clientID, tenantID); err != nil {
		if err.Error() == "oauth client does not belong to tenant" {
			middleware.RespondWithError(c, http.StatusForbidden, "forbidden",
				"Access denied to this OAuth client", nil)
			return
		}
		middleware.RespondWithError(c, http.StatusInternalServerError, "deletion_failed",
			"Failed to delete OAuth client", nil)
		return
	}

	c.Status(http.StatusNoContent)
}
