package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/identity/user"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService user.ServiceInterface
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService user.ServiceInterface) *UserHandler {
	return &UserHandler{userService: userService}
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

	c.JSON(http.StatusCreated, u)
}

// GetByID handles GET /api/v1/users/:id
func (h *UserHandler) GetByID(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

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

	// Verify user belongs to tenant
	if u.TenantID != tenantID {
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

	c.JSON(http.StatusOK, u)
}

// Delete handles DELETE /api/v1/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	// Get tenant ID from context
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid user ID format", nil)
		return
	}

	// Verify user belongs to tenant before deleting
	existingUser, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"User not found", nil)
		return
	}
	if existingUser.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
			"User does not belong to this tenant", nil)
		return
	}

	if err := h.userService.Delete(c.Request.Context(), id); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "delete_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

