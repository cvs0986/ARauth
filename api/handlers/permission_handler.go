package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/identity/permission"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// PermissionHandler handles permission-related HTTP requests
type PermissionHandler struct {
	permissionService permission.ServiceInterface
}

// NewPermissionHandler creates a new permission handler
func NewPermissionHandler(permissionService permission.ServiceInterface) *PermissionHandler {
	return &PermissionHandler{permissionService: permissionService}
}

// Create handles POST /api/v1/permissions
func (h *PermissionHandler) Create(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	var req permission.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Set tenant ID from context
	req.TenantID = tenantID

	createdPermission, err := h.permissionService.Create(c.Request.Context(), &req)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "creation_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusCreated, createdPermission)
}

// GetByID handles GET /api/v1/permissions/:id
func (h *PermissionHandler) GetByID(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid permission ID format", nil)
		return
	}

	perm, err := h.permissionService.GetByID(c.Request.Context(), id)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "permission_not_found",
			err.Error(), nil)
		return
	}

	// Verify permission belongs to tenant
	if perm.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
			"Permission does not belong to this tenant", nil)
		return
	}

	c.JSON(http.StatusOK, perm)
}

// Update handles PUT /api/v1/permissions/:id
func (h *PermissionHandler) Update(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid permission ID format", nil)
		return
	}

	// Verify permission belongs to tenant
	existingPermission, err := h.permissionService.GetByID(c.Request.Context(), id)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "permission_not_found",
			err.Error(), nil)
		return
	}
	if existingPermission.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
			"Permission does not belong to this tenant", nil)
		return
	}

	var req permission.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	updatedPermission, err := h.permissionService.Update(c.Request.Context(), id, &req)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "update_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, updatedPermission)
}

// Delete handles DELETE /api/v1/permissions/:id
func (h *PermissionHandler) Delete(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid permission ID format", nil)
		return
	}

	// Verify permission belongs to tenant
	existingPermission, err := h.permissionService.GetByID(c.Request.Context(), id)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "permission_not_found",
			err.Error(), nil)
		return
	}
	if existingPermission.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
			"Permission does not belong to this tenant", nil)
		return
	}

	if err := h.permissionService.Delete(c.Request.Context(), id); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "delete_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission deleted successfully"})
}

// List handles GET /api/v1/permissions
func (h *PermissionHandler) List(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	filters := &interfaces.PermissionFilters{
		Page:     1,
		PageSize: 20,
	}

	// Parse query parameters
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			filters.PageSize = pageSize
		}
	}

	if resource := c.Query("resource"); resource != "" {
		filters.Resource = &resource
	}

	if action := c.Query("action"); action != "" {
		filters.Action = &action
	}

	if search := c.Query("search"); search != "" {
		filters.Search = &search
	}

	permissions, err := h.permissionService.List(c.Request.Context(), tenantID, filters)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "list_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": permissions,
		"page":        filters.Page,
		"page_size":   filters.PageSize,
	})
}

// ListSystem handles GET /system/permissions - List all system permissions
// Note: This is a simplified implementation. System permissions are predefined.
// For now, we return an empty list and the frontend can handle this appropriately.
func (h *PermissionHandler) ListSystem(c *gin.Context) {
	// System permissions are predefined
	// Since the current service requires tenantID, we'll return an empty list for now
	// TODO: Modify repository/service to support listing system permissions without tenant_id
	c.JSON(http.StatusOK, gin.H{
		"permissions": []interface{}{},
		"page":        1,
		"page_size":   100,
		"message":     "System permissions are predefined. Please select a tenant to view permissions.",
	})
}

