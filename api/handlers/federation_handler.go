package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/auth/federation"
)

// FederationHandler handles federation-related HTTP requests
type FederationHandler struct {
	federationService federation.ServiceInterface
}

// NewFederationHandler creates a new federation handler
func NewFederationHandler(federationService federation.ServiceInterface) *FederationHandler {
	return &FederationHandler{
		federationService: federationService,
	}
}

// CreateIdentityProvider handles POST /api/v1/identity-providers
func (h *FederationHandler) CreateIdentityProvider(c *gin.Context) {
	tenantID, exists := middleware.GetTenantID(c)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_required", "message": "Tenant ID is required"})
		return
	}

	var req federation.CreateIdPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	provider, err := h.federationService.CreateIdentityProvider(c.Request.Context(), tenantID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "creation_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, provider)
}

// GetIdentityProvider handles GET /api/v1/identity-providers/:id
func (h *FederationHandler) GetIdentityProvider(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid identity provider ID"})
		return
	}

	provider, err := h.federationService.GetIdentityProvider(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Identity provider not found"})
		return
	}

	c.JSON(http.StatusOK, provider)
}

// ListIdentityProviders handles GET /api/v1/identity-providers
func (h *FederationHandler) ListIdentityProviders(c *gin.Context) {
	tenantID, exists := middleware.GetTenantID(c)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_required", "message": "Tenant ID is required"})
		return
	}

	providers, err := h.federationService.GetIdentityProvidersByTenant(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, providers)
}

// UpdateIdentityProvider handles PUT /api/v1/identity-providers/:id
func (h *FederationHandler) UpdateIdentityProvider(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid identity provider ID"})
		return
	}

	var req federation.UpdateIdPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	provider, err := h.federationService.UpdateIdentityProvider(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, provider)
}

// DeleteIdentityProvider handles DELETE /api/v1/identity-providers/:id
func (h *FederationHandler) DeleteIdentityProvider(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid identity provider ID"})
		return
	}

	if err := h.federationService.DeleteIdentityProvider(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// InitiateOIDCLogin handles GET /api/v1/auth/oidc/:provider_id/initiate
func (h *FederationHandler) InitiateOIDCLogin(c *gin.Context) {
	tenantID, exists := middleware.GetTenantID(c)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_required", "message": "Tenant ID is required"})
		return
	}

	providerIDStr := c.Param("provider_id")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_provider_id", "message": "Invalid provider ID"})
		return
	}

	redirectURI := c.Query("redirect_uri")
	if redirectURI == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "redirect_uri_required", "message": "redirect_uri query parameter is required"})
		return
	}

	authURL, state, err := h.federationService.InitiateOIDCLogin(c.Request.Context(), tenantID, providerID, redirectURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "initiation_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authorization_url": authURL,
		"state":            state,
	})
}

// HandleOIDCCallback handles GET /api/v1/auth/oidc/:provider_id/callback
func (h *FederationHandler) HandleOIDCCallback(c *gin.Context) {
	providerIDStr := c.Param("provider_id")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_provider_id", "message": "Invalid provider ID"})
		return
	}

	code := c.Query("code")
	state := c.Query("state")
	redirectURI := c.Query("redirect_uri")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code_required", "message": "code query parameter is required"})
		return
	}
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "state_required", "message": "state query parameter is required"})
		return
	}
	if redirectURI == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "redirect_uri_required", "message": "redirect_uri query parameter is required"})
		return
	}

	loginResp, err := h.federationService.HandleOIDCCallback(c.Request.Context(), providerID, code, state, redirectURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "callback_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, loginResp)
}

// InitiateSAMLLogin handles GET /api/v1/auth/saml/:provider_id/initiate
func (h *FederationHandler) InitiateSAMLLogin(c *gin.Context) {
	tenantID, exists := middleware.GetTenantID(c)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_required", "message": "Tenant ID is required"})
		return
	}

	providerIDStr := c.Param("provider_id")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_provider_id", "message": "Invalid provider ID"})
		return
	}

	acsURL := c.Query("acs_url")
	if acsURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "acs_url_required", "message": "acs_url query parameter is required"})
		return
	}

	redirectURL, err := h.federationService.InitiateSAMLLogin(c.Request.Context(), tenantID, providerID, acsURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "initiation_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"redirect_url": redirectURL,
	})
}

// HandleSAMLCallback handles POST /api/v1/auth/saml/:provider_id/callback
func (h *FederationHandler) HandleSAMLCallback(c *gin.Context) {
	providerIDStr := c.Param("provider_id")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_provider_id", "message": "Invalid provider ID"})
		return
	}

	var req struct {
		SAMLResponse string `json:"SAMLResponse" binding:"required"`
		RelayState   string `json:"RelayState,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	loginResp, err := h.federationService.HandleSAMLCallback(c.Request.Context(), providerID, req.SAMLResponse, req.RelayState)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "callback_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, loginResp)
}

