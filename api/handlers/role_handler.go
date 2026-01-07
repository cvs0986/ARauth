package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nuage-identity/iam/api/middleware"
	"github.com/nuage-identity/iam/identity/role"
	"github.com/nuage-identity/iam/storage/interfaces"
)

// RoleHandler handles role-related HTTP requests
type RoleHandler struct {
	roleService role.ServiceInterface
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(roleService role.ServiceInterface) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

// Create handles POST /api/v1/roles
func (h *RoleHandler) Create(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	var req role.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Set tenant ID from context
	req.TenantID = tenantID

	createdRole, err := h.roleService.Create(c.Request.Context(), &req)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "creation_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusCreated, createdRole)
}

// GetByID handles GET /api/v1/roles/:id
func (h *RoleHandler) GetByID(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid role ID format", nil)
		return
	}

	role, err := h.roleService.GetByID(c.Request.Context(), id)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "role_not_found",
			err.Error(), nil)
		return
	}

	// Verify role belongs to tenant
	if role.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
			"Role does not belong to this tenant", nil)
		return
	}

	c.JSON(http.StatusOK, role)
}

// Update handles PUT /api/v1/roles/:id
func (h *RoleHandler) Update(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid role ID format", nil)
		return
	}

	// Verify role belongs to tenant
	existingRole, err := h.roleService.GetByID(c.Request.Context(), id)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "role_not_found",
			err.Error(), nil)
		return
	}
	if existingRole.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
			"Role does not belong to this tenant", nil)
		return
	}

	var req role.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	updatedRole, err := h.roleService.Update(c.Request.Context(), id, &req)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "update_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, updatedRole)
}

// Delete handles DELETE /api/v1/roles/:id
func (h *RoleHandler) Delete(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid role ID format", nil)
		return
	}

	// Verify role belongs to tenant
	existingRole, err := h.roleService.GetByID(c.Request.Context(), id)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "role_not_found",
			err.Error(), nil)
		return
	}
	if existingRole.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
			"Role does not belong to this tenant", nil)
		return
	}

	if err := h.roleService.Delete(c.Request.Context(), id); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "delete_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}

// List handles GET /api/v1/roles
func (h *RoleHandler) List(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	filters := &interfaces.RoleFilters{
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

	if isSystemStr := c.Query("is_system"); isSystemStr != "" {
		isSystem := isSystemStr == "true"
		filters.IsSystem = &isSystem
	}

	if search := c.Query("search"); search != "" {
		filters.Search = &search
	}

	roles, err := h.roleService.List(c.Request.Context(), tenantID, filters)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "list_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles":    roles,
		"page":     filters.Page,
		"page_size": filters.PageSize,
	})
}

// GetUserRoles handles GET /api/v1/users/:user_id/roles
func (h *RoleHandler) GetUserRoles(c *gin.Context) {
	// Get tenant ID from context
	_, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_user_id",
			"Invalid user ID format", nil)
		return
	}

	roles, err := h.roleService.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "list_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// AssignRoleToUser handles POST /api/v1/users/:user_id/roles/:role_id
func (h *RoleHandler) AssignRoleToUser(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_user_id",
			"Invalid user ID format", nil)
		return
	}

	roleIDStr := c.Param("role_id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_role_id",
			"Invalid role ID format", nil)
		return
	}

	// Verify role belongs to tenant
	existingRole, err := h.roleService.GetByID(c.Request.Context(), roleID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "role_not_found",
			err.Error(), nil)
		return
	}
	if existingRole.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
			"Role does not belong to this tenant", nil)
		return
	}

	if err := h.roleService.AssignRoleToUser(c.Request.Context(), userID, roleID); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "assignment_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role assigned successfully"})
}

// RemoveRoleFromUser handles DELETE /api/v1/users/:user_id/roles/:role_id
func (h *RoleHandler) RemoveRoleFromUser(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_user_id",
			"Invalid user ID format", nil)
		return
	}

	roleIDStr := c.Param("role_id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_role_id",
			"Invalid role ID format", nil)
		return
	}

	// Verify role belongs to tenant
	existingRole, err := h.roleService.GetByID(c.Request.Context(), roleID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "role_not_found",
			err.Error(), nil)
		return
	}
	if existingRole.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
			"Role does not belong to this tenant", nil)
		return
	}

	if err := h.roleService.RemoveRoleFromUser(c.Request.Context(), userID, roleID); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "removal_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role removed successfully"})
}

// GetRolePermissions handles GET /api/v1/roles/:id/permissions
func (h *RoleHandler) GetRolePermissions(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	roleIDStr := c.Param("id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_role_id",
			"Invalid role ID format", nil)
		return
	}

	// Verify role belongs to tenant
	existingRole, err := h.roleService.GetByID(c.Request.Context(), roleID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "role_not_found",
			err.Error(), nil)
		return
	}
	if existingRole.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
			"Role does not belong to this tenant", nil)
		return
	}

	permissions, err := h.roleService.GetRolePermissions(c.Request.Context(), roleID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "list_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

// AssignPermissionToRole handles POST /api/v1/roles/:id/permissions/:permission_id
func (h *RoleHandler) AssignPermissionToRole(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	roleIDStr := c.Param("id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_role_id",
			"Invalid role ID format", nil)
		return
	}

	permissionIDStr := c.Param("permission_id")
	permissionID, err := uuid.Parse(permissionIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_permission_id",
			"Invalid permission ID format", nil)
		return
	}

	// Verify role belongs to tenant
	existingRole, err := h.roleService.GetByID(c.Request.Context(), roleID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "role_not_found",
			err.Error(), nil)
		return
	}
	if existingRole.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
			"Role does not belong to this tenant", nil)
		return
	}

	if err := h.roleService.AssignPermissionToRole(c.Request.Context(), roleID, permissionID); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "assignment_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission assigned successfully"})
}

// RemovePermissionFromRole handles DELETE /api/v1/roles/:id/permissions/:permission_id
func (h *RoleHandler) RemovePermissionFromRole(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	roleIDStr := c.Param("id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_role_id",
			"Invalid role ID format", nil)
		return
	}

	permissionIDStr := c.Param("permission_id")
	permissionID, err := uuid.Parse(permissionIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_permission_id",
			"Invalid permission ID format", nil)
		return
	}

	// Verify role belongs to tenant
	existingRole, err := h.roleService.GetByID(c.Request.Context(), roleID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "role_not_found",
			err.Error(), nil)
		return
	}
	if existingRole.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
			"Role does not belong to this tenant", nil)
		return
	}

	if err := h.roleService.RemovePermissionFromRole(c.Request.Context(), roleID, permissionID); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "removal_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission removed successfully"})
}

