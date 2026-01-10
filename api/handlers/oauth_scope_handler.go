package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/identity/audit"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/identity/oauth_scope"
)

// OAuthScopeHandler handles OAuth scope-related HTTP requests
type OAuthScopeHandler struct {
	scopeService oauth_scope.ServiceInterface
	auditService audit.ServiceInterface
}

// NewOAuthScopeHandler creates a new OAuth scope handler
func NewOAuthScopeHandler(scopeService oauth_scope.ServiceInterface, auditService audit.ServiceInterface) *OAuthScopeHandler {
	return &OAuthScopeHandler{
		scopeService: scopeService,
		auditService: auditService,
	}
}

// CreateScope handles POST /api/v1/oauth/scopes
func (h *OAuthScopeHandler) CreateScope(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	var req oauth_scope.CreateScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Create scope
	scope, err := h.scopeService.CreateScope(c.Request.Context(), tenantID, &req)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "creation_failed",
			err.Error(), nil)
		return
	}

	// Log audit event
	actor, _ := extractActorFromContext(c)
	sourceIP, userAgent := extractSourceInfo(c)
	target := &models.AuditTarget{
		Type:       "oauth_scope",
		ID:         scope.ID,
		Identifier: scope.Name,
	}
	event := &models.AuditEvent{
		EventType: models.EventTypeOAuthScopeCreated,
		Actor:     actor,
		Target:    target,
		TenantID:  &tenantID,
		SourceIP:  sourceIP,
		UserAgent: userAgent,
		Metadata: map[string]interface{}{
			"scope_name": scope.Name,
			"permissions": scope.Permissions,
			"is_default": scope.IsDefault,
		},
		Result: models.ResultSuccess,
	}
	event.Flatten()
	_ = h.auditService.LogEvent(c.Request.Context(), event)

	c.JSON(http.StatusCreated, scope)
}

// GetScope handles GET /api/v1/oauth/scopes/:id
func (h *OAuthScopeHandler) GetScope(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	scopeIDStr := c.Param("id")
	scopeID, err := uuid.Parse(scopeIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid scope ID format", nil)
		return
	}

	// Get scope
	scope, err := h.scopeService.GetScope(c.Request.Context(), scopeID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"OAuth scope not found", nil)
		return
	}

	// Verify scope belongs to tenant
	if scope.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
			"Scope does not belong to this tenant", nil)
		return
	}

	c.JSON(http.StatusOK, scope)
}

// ListScopes handles GET /api/v1/oauth/scopes
func (h *OAuthScopeHandler) ListScopes(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	// Parse filters
	var filters oauth_scope.ScopeFilters
	filters.Page = 1
	filters.PageSize = 20

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			if pageSize > 100 {
				pageSize = 100
			}
			filters.PageSize = pageSize
		}
	}

	if isDefaultStr := c.Query("is_default"); isDefaultStr != "" {
		isDefault := isDefaultStr == "true"
		filters.IsDefault = &isDefault
	}

	// List scopes
	scopes, err := h.scopeService.ListScopes(c.Request.Context(), tenantID, &filters)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "query_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"scopes": scopes,
		"total":  len(scopes),
	})
}

// UpdateScope handles PUT /api/v1/oauth/scopes/:id
func (h *OAuthScopeHandler) UpdateScope(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	scopeIDStr := c.Param("id")
	scopeID, err := uuid.Parse(scopeIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid scope ID format", nil)
		return
	}

	var req oauth_scope.UpdateScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Get existing scope to verify tenant ownership
	existingScope, err := h.scopeService.GetScope(c.Request.Context(), scopeID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"OAuth scope not found", nil)
		return
	}

	if existingScope.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
			"Scope does not belong to this tenant", nil)
		return
	}

	// Update scope
	scope, err := h.scopeService.UpdateScope(c.Request.Context(), scopeID, &req)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "update_failed",
			err.Error(), nil)
		return
	}

	// Log audit event
	actor, _ := extractActorFromContext(c)
	sourceIP, userAgent := extractSourceInfo(c)
	target := &models.AuditTarget{
		Type:       "oauth_scope",
		ID:         scope.ID,
		Identifier: scope.Name,
	}
	event := &models.AuditEvent{
		EventType: models.EventTypeOAuthScopeUpdated,
		Actor:     actor,
		Target:    target,
		TenantID:  &tenantID,
		SourceIP:  sourceIP,
		UserAgent: userAgent,
		Metadata: map[string]interface{}{
			"scope_name": scope.Name,
			"permissions": scope.Permissions,
			"is_default": scope.IsDefault,
		},
		Result: models.ResultSuccess,
	}
	event.Flatten()
	_ = h.auditService.LogEvent(c.Request.Context(), event)

	c.JSON(http.StatusOK, scope)
}

// DeleteScope handles DELETE /api/v1/oauth/scopes/:id
func (h *OAuthScopeHandler) DeleteScope(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	scopeIDStr := c.Param("id")
	scopeID, err := uuid.Parse(scopeIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid scope ID format", nil)
		return
	}

	// Get existing scope to verify tenant ownership and get name for audit
	existingScope, err := h.scopeService.GetScope(c.Request.Context(), scopeID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"OAuth scope not found", nil)
		return
	}

	if existingScope.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
			"Scope does not belong to this tenant", nil)
		return
	}

	// Delete scope
	err = h.scopeService.DeleteScope(c.Request.Context(), scopeID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "deletion_failed",
			err.Error(), nil)
		return
	}

	// Log audit event
	actor, _ := extractActorFromContext(c)
	sourceIP, userAgent := extractSourceInfo(c)
	target := &models.AuditTarget{
		Type:       "oauth_scope",
		ID:         scopeID,
		Identifier: existingScope.Name,
	}
	event := &models.AuditEvent{
		EventType: models.EventTypeOAuthScopeDeleted,
		Actor:     actor,
		Target:    target,
		TenantID:  &tenantID,
		SourceIP:  sourceIP,
		UserAgent: userAgent,
		Result:    models.ResultSuccess,
	}
	event.Flatten()
	_ = h.auditService.LogEvent(c.Request.Context(), event)

	c.JSON(http.StatusOK, gin.H{
		"message": "OAuth scope deleted",
	})
}
