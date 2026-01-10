package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/identity/audit"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/identity/role"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// RoleHandler handles role-related HTTP requests
type RoleHandler struct {
	roleService     role.ServiceInterface
	systemRoleRepo  interfaces.SystemRoleRepository
	userRepo        interfaces.UserRepository
	auditService    audit.ServiceInterface
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(roleService role.ServiceInterface, systemRoleRepo interfaces.SystemRoleRepository, userRepo interfaces.UserRepository, auditService audit.ServiceInterface) *RoleHandler {
	return &RoleHandler{
		roleService:    roleService,
		systemRoleRepo: systemRoleRepo,
		userRepo:       userRepo,
		auditService:   auditService,
	}
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

	// Log audit event
	if actor, err := extractActorFromContext(c); err == nil {
		sourceIP, userAgent := extractSourceInfo(c)
		target := &models.AuditTarget{
			Type:       "role",
			ID:         createdRole.ID,
			Identifier: createdRole.Name,
		}
		_ = h.auditService.LogRoleCreated(c.Request.Context(), actor, target, &tenantID, sourceIP, userAgent)
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

	// Log audit event
	if actor, err := extractActorFromContext(c); err == nil {
		sourceIP, userAgent := extractSourceInfo(c)
		target := &models.AuditTarget{
			Type:       "role",
			ID:         updatedRole.ID,
			Identifier: updatedRole.Name,
		}
		_ = h.auditService.LogRoleUpdated(c.Request.Context(), actor, target, &tenantID, sourceIP, userAgent)
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

	// Log audit event
	if actor, err := extractActorFromContext(c); err == nil {
		sourceIP, userAgent := extractSourceInfo(c)
		target := &models.AuditTarget{
			Type:       "role",
			ID:         existingRole.ID,
			Identifier: existingRole.Name,
		}
		_ = h.auditService.LogRoleDeleted(c.Request.Context(), actor, target, &tenantID, sourceIP, userAgent)
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

// ListSystem handles GET /system/roles - List all system roles
func (h *RoleHandler) ListSystem(c *gin.Context) {
	// Fetch system roles from system_roles table
	systemRoles, err := h.systemRoleRepo.List(c.Request.Context())
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "list_failed",
			err.Error(), nil)
		return
	}

	// Convert SystemRole to Role format for API response
	roles := make([]*models.Role, 0, len(systemRoles))
	for _, sr := range systemRoles {
		createdAt := parseTime(sr.CreatedAt)
		updatedAt := parseTime(sr.UpdatedAt)
		
		// Handle empty description
		var description *string
		if sr.Description != "" {
			description = &sr.Description
		}
		
		role := &models.Role{
			ID:          sr.ID,
			Name:        sr.Name,
			Description: description,
			IsSystem:    true,
			TenantID:    uuid.Nil, // System roles don't have tenant_id
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		}
		roles = append(roles, role)
	}

	c.JSON(http.StatusOK, gin.H{
		"roles":     roles,
		"page":      1,
		"page_size": 100,
		"total":     len(roles),
	})
}

// parseTime parses a time string in RFC3339 format
func parseTime(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		// If parsing fails, try other common formats
		t, err = time.Parse("2006-01-02 15:04:05", timeStr)
		if err != nil {
			return time.Time{}
		}
	}
	return t
}

// GetUserRoles handles GET /api/v1/users/:id/roles
func (h *RoleHandler) GetUserRoles(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_user_id",
			"Invalid user ID format", nil)
		return
	}

	// Check if user is a system user
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "user_not_found",
			"User not found", nil)
		return
	}

	var roles []*models.Role

	// If system user, fetch system roles
	if user.PrincipalType == models.PrincipalTypeSystem {
		systemRoles, err := h.systemRoleRepo.GetUserSystemRoles(c.Request.Context(), userID)
		if err != nil {
			middleware.RespondWithError(c, http.StatusInternalServerError, "list_failed",
				err.Error(), nil)
			return
		}

		// Convert system roles to models.Role
		roles = make([]*models.Role, len(systemRoles))
		for i, sr := range systemRoles {
			var desc *string
			if sr.Description != "" {
				desc = &sr.Description
			}
			createdAt, _ := time.Parse(time.RFC3339, sr.CreatedAt)
			updatedAt, _ := time.Parse(time.RFC3339, sr.UpdatedAt)
			roles[i] = &models.Role{
				ID:          sr.ID,
				Name:        sr.Name,
				Description: desc,
				IsSystem:    true,
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			}
		}
	} else {
		// Tenant user - require tenant context
		_, ok := middleware.RequireTenant(c)
		if !ok {
			return
		}

		tenantRoles, err := h.roleService.GetUserRoles(c.Request.Context(), userID)
		if err != nil {
			middleware.RespondWithError(c, http.StatusInternalServerError, "list_failed",
				err.Error(), nil)
			return
		}
		roles = tenantRoles
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// AssignRoleToUser handles POST /api/v1/users/:id/roles/:role_id
func (h *RoleHandler) AssignRoleToUser(c *gin.Context) {
	userIDStr := c.Param("id")
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

	// Check if this is a system role by trying to get it from system roles
	systemRole, err := h.systemRoleRepo.GetByID(c.Request.Context(), roleID)
	if err == nil && systemRole != nil {
		// This is a system role - assign using SystemRoleRepository
		// System roles don't require tenant context
		
		// Get current user's system roles from JWT claims for permission checks
		claimsObj, exists := c.Get("user_claims")
		if !exists {
			middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
				"User claims not found", nil)
			return
		}

		userClaims := claimsObj.(*claims.Claims)

		// Check if trying to assign system_owner role - only system_owner can assign this
		if systemRole.Name == "system_owner" {
			hasOwnerRole := false
			for _, roleName := range userClaims.SystemRoles {
				if roleName == "system_owner" {
					hasOwnerRole = true
					break
				}
			}
			if !hasOwnerRole {
				middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
					"Only system_owner can assign system_owner role", nil)
				return
			}
		}

		// Check if trying to assign system_auditor role - only system_owner can assign this
		if systemRole.Name == "system_auditor" {
			hasOwnerRole := false
			for _, roleName := range userClaims.SystemRoles {
				if roleName == "system_owner" {
					hasOwnerRole = true
					break
				}
			}
			if !hasOwnerRole {
				middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
					"Only system_owner can assign system_auditor role", nil)
				return
			}
		}

		// Get current user ID from JWT for assigned_by
		var assignedBy *uuid.UUID
		if userClaims.Subject != "" {
			if parsedID, err := uuid.Parse(userClaims.Subject); err == nil {
				assignedBy = &parsedID
			}
		}
		
		if err := h.systemRoleRepo.AssignRoleToUser(c.Request.Context(), userID, roleID, assignedBy); err != nil {
			middleware.RespondWithError(c, http.StatusBadRequest, "assignment_failed",
				err.Error(), nil)
			return
		}

		// Log audit event for system role assignment
		if actor, err := extractActorFromContext(c); err == nil {
			sourceIP, userAgent := extractSourceInfo(c)
			// Get user info for target
			user, err := h.userRepo.GetByID(c.Request.Context(), userID)
			if err == nil {
				target := &models.AuditTarget{
					Type:       "user",
					ID:         userID,
					Identifier: user.Username,
				}
				_ = h.auditService.LogRoleAssigned(c.Request.Context(), actor, target, nil, sourceIP, userAgent, map[string]interface{}{
					"role_id":   roleID.String(),
					"role_name": systemRole.Name,
					"is_system": true,
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "System role assigned successfully"})
		return
	}

	// This is a tenant role - require tenant context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
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

// RemoveRoleFromUser handles DELETE /api/v1/users/:id/roles/:role_id
func (h *RoleHandler) RemoveRoleFromUser(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	userIDStr := c.Param("id")
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

	// CRITICAL SECURITY: Prevent removal of last tenant_owner
	// This prevents tenant self-lockout scenarios
	// Invariant: At least one user must always have tenant_owner role
	if existingRole.Name == "tenant_owner" {
		// Count how many users currently have tenant_owner role
		allUsers, err := h.userRepo.List(c.Request.Context(), tenantID, &interfaces.UserFilters{
			Page:     1,
			PageSize: 1000, // Get all users to count tenant_owners
		})
		if err == nil {
			tenantOwnerCount := 0
			for _, u := range allUsers {
				userRoles, err := h.roleService.GetUserRoles(c.Request.Context(), u.ID)
				if err == nil {
					for _, ur := range userRoles {
						if ur.ID == roleID {
							tenantOwnerCount++
							break
						}
					}
				}
			}
			
			// If this is the only tenant_owner, prevent removal
			if tenantOwnerCount <= 1 {
				middleware.RespondWithError(c, http.StatusForbidden, "cannot_remove_last_owner",
					"Cannot remove tenant_owner role from the last user with this role. At least one user must always have tenant_owner role to prevent tenant lockout. Use break-glass procedure if absolutely necessary.", nil)
				return
			}
		}
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

