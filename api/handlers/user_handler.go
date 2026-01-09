package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/identity/audit"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/identity/user"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService    user.ServiceInterface
	systemRoleRepo interfaces.SystemRoleRepository
	roleRepo       interfaces.RoleRepository
	auditService   audit.ServiceInterface
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService user.ServiceInterface, systemRoleRepo interfaces.SystemRoleRepository, roleRepo interfaces.RoleRepository, auditService audit.ServiceInterface) *UserHandler {
	return &UserHandler{
		userService:    userService,
		systemRoleRepo: systemRoleRepo,
		roleRepo:       roleRepo,
		auditService:   auditService,
	}
}

// Create handles POST /api/v1/users
func (h *UserHandler) Create(c *gin.Context) {
	// Get tenant ID from context (set by tenant middleware)
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	var req user.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Set tenant ID from context (always override any tenant_id in request body for security)
	req.TenantID = tenantID

	u, err := h.userService.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "creation_failed",
			"message": err.Error(),
		})
		return
	}

	// Check if this is the first user in the tenant - assign tenant_owner role
	users, err := h.userService.List(c.Request.Context(), tenantID, &interfaces.UserFilters{
		Page:     1,
		PageSize: 2, // We only need to check if there are 2 or fewer users
	})
	if err == nil && len(users) <= 1 {
		// This is the first user - assign tenant_owner role
		tenantOwnerRole, err := h.roleRepo.GetByName(c.Request.Context(), tenantID, "tenant_owner")
		if err == nil && tenantOwnerRole != nil {
			// Assign tenant_owner role to the new user
			if assignErr := h.roleRepo.AssignRoleToUser(c.Request.Context(), u.ID, tenantOwnerRole.ID); assignErr != nil {
				// Log error but don't fail user creation
				// The role can be assigned manually later if needed
				_ = assignErr
			}
		}
	}

	// Log audit event
	if actor, err := extractActorFromContext(c); err == nil {
		sourceIP, userAgent := extractSourceInfo(c)
		target := &audit.AuditTarget{
			Type:       "user",
			ID:         u.ID,
			Identifier: u.Username,
		}
		_ = h.auditService.LogUserCreated(c.Request.Context(), actor, target, &tenantID, sourceIP, userAgent, map[string]interface{}{
			"email":      u.Email,
			"first_name": u.FirstName,
			"last_name":  u.LastName,
		})
	}

	c.JSON(http.StatusCreated, u)
}

// CreateSystem handles POST /api/v1/system/users
func (h *UserHandler) CreateSystem(c *gin.Context) {
	// Verify user is SYSTEM user
	principalType, exists := c.Get("principal_type")
	if !exists || principalType != "SYSTEM" {
		middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
			"Only SYSTEM users can create system users", nil)
		return
	}

	// Get current user's system roles from JWT claims
	claimsObj, exists := c.Get("user_claims")
	if !exists {
		middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
			"User claims not found", nil)
		return
	}

	userClaims := claimsObj.(*claims.Claims)

	// Check if current user is system_auditor (read-only, cannot create)
	for _, roleName := range userClaims.SystemRoles {
		if roleName == "system_auditor" {
			middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
				"system_auditor has read-only access and cannot create users", nil)
			return
		}
	}

	// Check if current user has system:users permission
	hasSystemUsersPermission := false
	for _, perm := range userClaims.SystemPermissions {
		if perm == "system:users" || perm == "system:*" || perm == "*:*" {
			hasSystemUsersPermission = true
			break
		}
	}

	if !hasSystemUsersPermission {
		middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
			"Required permission: system:users", nil)
		return
	}

	// No tenant required for system users
	var req user.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Ensure tenant_id is not set (system users don't have tenant_id)
	req.TenantID = uuid.Nil

	u, err := h.userService.CreateSystem(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "creation_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, u)
}

// GetByID handles GET /api/v1/users/:id
func (h *UserHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid user ID format", nil)
		return
	}

	u, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"User not found", nil)
		return
	}

	// Check if this is a system user
	if u.PrincipalType == models.PrincipalTypeSystem {
		// System users don't require tenant context
		// Verify the requesting user is a SYSTEM user
		principalType, exists := c.Get("principal_type")
		if !exists || principalType != "SYSTEM" {
			middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
				"Only SYSTEM users can access system user details", nil)
			return
		}
		c.JSON(http.StatusOK, u)
		return
	}

	// This is a tenant user - require tenant context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	// Verify user belongs to tenant
	if u.TenantID == nil || *u.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
			"User does not belong to this tenant", nil)
		return
	}

	c.JSON(http.StatusOK, u)
}

// List handles GET /api/v1/users
func (h *UserHandler) List(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize > 100 {
		pageSize = 100
	}

	// Parse filters
	filters := &interfaces.UserFilters{
		Page:     page,
		PageSize: pageSize,
	}

	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}

	if search := c.Query("search"); search != "" {
		filters.Search = &search
	}

	// Get users
	users, err := h.userService.List(c.Request.Context(), tenantID, filters)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "list_failed",
			err.Error(), nil)
		return
	}

	// Get total count
	totalCount, err := h.userService.Count(c.Request.Context(), tenantID, filters)
	total := int64(totalCount)
	if err != nil {
		total = int64(len(users)) // Fallback
	}

	c.JSON(http.StatusOK, gin.H{
		"users":     users,
		"page":      filters.Page,
		"page_size": filters.PageSize,
		"total":     total,
	})
}

// ListSystem handles GET /system/users - List all system users (system admin only)
func (h *UserHandler) ListSystem(c *gin.Context) {
	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize > 100 {
		pageSize = 100
	}

	// Parse filters
	filters := &interfaces.UserFilters{
		Page:     page,
		PageSize: pageSize,
	}

	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}

	if search := c.Query("search"); search != "" {
		filters.Search = &search
	}

	// Get system users
	users, err := h.userService.ListSystem(c.Request.Context(), filters)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "list_failed",
			err.Error(), nil)
		return
	}

	// Get total count
	totalCount, err := h.userService.CountSystem(c.Request.Context(), filters)
	total := int64(totalCount)
	if err != nil {
		total = int64(len(users)) // Fallback
	}

	c.JSON(http.StatusOK, gin.H{
		"users":     users,
		"page":      filters.Page,
		"page_size": filters.PageSize,
		"total":     total,
	})
}

// Update handles PUT /api/v1/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "Invalid user ID format",
		})
		return
	}

	// Get the user being updated
	existingUser, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"User not found", nil)
		return
	}

	// Check if this is a system user
	if existingUser.PrincipalType == models.PrincipalTypeSystem {
		// System user update - check permissions
		principalType, exists := c.Get("principal_type")
		if !exists || principalType != "SYSTEM" {
			middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
				"Only SYSTEM users can update system users", nil)
			return
		}

		// Get current user's system roles from JWT claims
		claimsObj, exists := c.Get("user_claims")
		if !exists {
			middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
				"User claims not found", nil)
			return
		}

		userClaims := claimsObj.(*claims.Claims)

		// Check if current user is system_auditor (read-only, cannot update)
		for _, roleName := range userClaims.SystemRoles {
			if roleName == "system_auditor" {
				middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
					"system_auditor has read-only access and cannot update users", nil)
				return
			}
		}

		// Check if target user has system_owner role
		targetUserRoles, err := h.systemRoleRepo.GetUserSystemRoles(c.Request.Context(), id)
		if err == nil {
			for _, role := range targetUserRoles {
				if role.Name == "system_owner" {
					// Check if current user is system_owner (only system_owner can edit system_owner)
					hasOwnerRole := false
					for _, roleName := range userClaims.SystemRoles {
						if roleName == "system_owner" {
							hasOwnerRole = true
							break
						}
					}
					if !hasOwnerRole {
						middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
							"Only system_owner can update system_owner", nil)
						return
					}
				}
				if role.Name == "system_auditor" {
					// system_auditor cannot be updated by anyone except system_owner
					hasOwnerRole := false
					for _, roleName := range userClaims.SystemRoles {
						if roleName == "system_owner" {
							hasOwnerRole = true
							break
						}
					}
					if !hasOwnerRole {
						middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
							"Only system_owner can update system_auditor", nil)
						return
					}
				}
			}
		}
	} else {
		// Tenant user update - require tenant context
		tenantID, ok := middleware.RequireTenant(c)
		if !ok {
			return
		}

		// Verify user belongs to tenant
		if existingUser.TenantID == nil || *existingUser.TenantID != tenantID {
			middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
				"User does not belong to this tenant", nil)
			return
		}
	}

	var req user.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	u, err := h.userService.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "update_failed",
			"message": err.Error(),
		})
		return
	}

	// Log audit event
	if actor, err := extractActorFromContext(c); err == nil {
		sourceIP, userAgent := extractSourceInfo(c)
		target := &audit.AuditTarget{
			Type:       "user",
			ID:         u.ID,
			Identifier: u.Username,
		}
		var tenantID *uuid.UUID
		if u.TenantID != nil {
			tenantID = u.TenantID
		}
		_ = h.auditService.LogUserUpdated(c.Request.Context(), actor, target, tenantID, sourceIP, userAgent, map[string]interface{}{
			"email":      u.Email,
			"first_name": u.FirstName,
			"last_name":  u.LastName,
		})
	}

	c.JSON(http.StatusOK, u)
}

// Delete handles DELETE /api/v1/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid user ID format", nil)
		return
	}

	// Get the user being deleted
	existingUser, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"User not found", nil)
		return
	}

	// Check if this is a system user
	if existingUser.PrincipalType == models.PrincipalTypeSystem {
		// System user deletion - check permissions
		principalType, exists := c.Get("principal_type")
		if !exists || principalType != "SYSTEM" {
			middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
				"Only SYSTEM users can delete system users", nil)
			return
		}

		// Get current user's system roles from JWT claims
		claimsObj, exists := c.Get("user_claims")
		if !exists {
			middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
				"User claims not found", nil)
			return
		}

		// Check if target user has system_owner role
		targetUserRoles, err := h.systemRoleRepo.GetUserSystemRoles(c.Request.Context(), id)
		if err == nil {
			for _, role := range targetUserRoles {
				if role.Name == "system_owner" {
					// Only system_owner can delete system_owner (but typically shouldn't delete themselves)
					userClaims := claimsObj.(*claims.Claims)
					hasOwnerRole := false
					for _, roleName := range userClaims.SystemRoles {
						if roleName == "system_owner" {
							hasOwnerRole = true
							break
						}
					}
					if !hasOwnerRole {
						middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
							"Only system_owner can delete system_owner", nil)
						return
					}
				}
				if role.Name == "system_auditor" {
					// system_auditor cannot be deleted by anyone except system_owner
					userClaims := claimsObj.(*claims.Claims)
					hasOwnerRole := false
					for _, roleName := range userClaims.SystemRoles {
						if roleName == "system_owner" {
							hasOwnerRole = true
							break
						}
					}
					if !hasOwnerRole {
						middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
							"Only system_owner can delete system_auditor", nil)
						return
					}
				}
			}
		}

		// Check if current user is system_auditor (read-only, cannot delete)
		userClaims := claimsObj.(*claims.Claims)
		for _, roleName := range userClaims.SystemRoles {
			if roleName == "system_auditor" {
				middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
					"system_auditor has read-only access and cannot delete users", nil)
				return
			}
		}
	} else {
		// Tenant user deletion - require tenant context
		tenantID, ok := middleware.RequireTenant(c)
		if !ok {
			return
		}

		// Verify user belongs to tenant
		if existingUser.TenantID == nil || *existingUser.TenantID != tenantID {
			middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
				"User does not belong to this tenant", nil)
			return
		}
	}

	if err := h.userService.Delete(c.Request.Context(), id); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "delete_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// GetUserPermissions handles GET /api/v1/users/:id/permissions
func (h *UserHandler) GetUserPermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid user ID format", nil)
		return
	}

	u, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"User not found", nil)
		return
	}

	var permissions []interface{}

	// If system user, fetch permissions from system roles
	if u.PrincipalType == models.PrincipalTypeSystem {
		// Get system roles for user
		systemRoles, err := h.systemRoleRepo.GetUserSystemRoles(c.Request.Context(), id)
		if err != nil {
			middleware.RespondWithError(c, http.StatusInternalServerError, "list_failed",
				err.Error(), nil)
			return
		}

		// Aggregate permissions from all system roles
		permissionMap := make(map[string]bool)
		for _, role := range systemRoles {
			rolePerms, err := h.systemRoleRepo.GetRolePermissions(c.Request.Context(), role.ID)
			if err != nil {
				continue // Skip if we can't get permissions for this role
			}
			for _, perm := range rolePerms {
				key := perm.Resource + ":" + perm.Action
				permissionMap[key] = true
			}
		}

		// Convert to list
		permissions = make([]interface{}, 0, len(permissionMap))
		for key := range permissionMap {
			parts := strings.Split(key, ":")
			if len(parts) == 2 {
				permissions = append(permissions, gin.H{
					"resource":    parts[0],
					"action":      parts[1],
					"permission":  key,
					"description": fmt.Sprintf("%s %s", parts[0], parts[1]),
				})
			}
		}
	} else {
		// Tenant user - require tenant context
		tenantID, ok := middleware.RequireTenant(c)
		if !ok {
			return
		}

		// Verify user belongs to tenant
		if u.TenantID == nil || *u.TenantID != tenantID {
			middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
				"User does not belong to this tenant", nil)
			return
		}

		// TODO: Implement tenant user permissions aggregation
		// For now, return empty list
		permissions = []interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

