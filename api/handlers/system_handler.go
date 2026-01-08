package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/identity/tenant"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// SystemHandler handles system-level operations (tenant management, system settings)
type SystemHandler struct {
	tenantService      tenant.ServiceInterface
	tenantRepo         interfaces.TenantRepository
	tenantSettingsRepo interfaces.TenantSettingsRepository
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(tenantService tenant.ServiceInterface, tenantRepo interfaces.TenantRepository, tenantSettingsRepo interfaces.TenantSettingsRepository) *SystemHandler {
	return &SystemHandler{
		tenantService:      tenantService,
		tenantRepo:         tenantRepo,
		tenantSettingsRepo: tenantSettingsRepo,
	}
}

// ListTenants handles GET /system/tenants - List all tenants (system admin only)
func (h *SystemHandler) ListTenants(c *gin.Context) {
	// Get filters from query params
	filters := &interfaces.TenantFilters{
		Page:     1,
		PageSize: 20,
	}

	if page := c.Query("page"); page != "" {
		// Parse page (simplified, add proper parsing)
		_ = page // TODO: parse page number
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		_ = pageSize // TODO: parse page size
	}
	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}
	if search := c.Query("search"); search != "" {
		filters.Search = &search
	}

	tenants, err := h.tenantRepo.List(c.Request.Context(), filters)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to list tenants", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenants": tenants,
		"page":    filters.Page,
		"page_size": filters.PageSize,
	})
}

// GetTenant handles GET /system/tenants/:id - Get tenant by ID (system admin only)
func (h *SystemHandler) GetTenant(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid tenant ID", nil)
		return
	}

	tenant, err := h.tenantRepo.GetByID(c.Request.Context(), tenantID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"Tenant not found", nil)
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// CreateTenant handles POST /system/tenants - Create new tenant (system admin only)
func (h *SystemHandler) CreateTenant(c *gin.Context) {
	var req struct {
		Name   string                 `json:"name" binding:"required"`
		Domain string                 `json:"domain" binding:"required"`
		Status string                 `json:"status,omitempty"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	if req.Status == "" {
		req.Status = models.TenantStatusActive
	}

	tenant := &models.Tenant{
		Name:     req.Name,
		Domain:   req.Domain,
		Status:   req.Status,
		Metadata: req.Metadata,
	}

	if err := h.tenantRepo.Create(c.Request.Context(), tenant); err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to create tenant: "+err.Error(), nil)
		return
	}

	c.JSON(http.StatusCreated, tenant)
}

// UpdateTenant handles PUT /system/tenants/:id - Update tenant (system admin only)
func (h *SystemHandler) UpdateTenant(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid tenant ID", nil)
		return
	}

	var req struct {
		Name     *string                `json:"name,omitempty"`
		Domain   *string                `json:"domain,omitempty"`
		Status   *string                `json:"status,omitempty"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Get existing tenant
	existing, err := h.tenantRepo.GetByID(c.Request.Context(), tenantID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"Tenant not found", nil)
		return
	}

	// Update fields
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Domain != nil {
		existing.Domain = *req.Domain
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}
	if req.Metadata != nil {
		existing.Metadata = req.Metadata
	}

	if err := h.tenantRepo.Update(c.Request.Context(), existing); err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to update tenant: "+err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, existing)
}

// DeleteTenant handles DELETE /system/tenants/:id - Delete tenant (system admin only)
func (h *SystemHandler) DeleteTenant(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid tenant ID", nil)
		return
	}

	if err := h.tenantRepo.Delete(c.Request.Context(), tenantID); err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to delete tenant: "+err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tenant deleted successfully"})
}

// SuspendTenant handles POST /system/tenants/:id/suspend - Suspend tenant (system admin only)
func (h *SystemHandler) SuspendTenant(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid tenant ID", nil)
		return
	}

	existing, err := h.tenantRepo.GetByID(c.Request.Context(), tenantID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"Tenant not found", nil)
		return
	}

	existing.Status = models.TenantStatusSuspended
	if err := h.tenantRepo.Update(c.Request.Context(), existing); err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to suspend tenant: "+err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, existing)
}

// ResumeTenant handles POST /system/tenants/:id/resume - Resume tenant (system admin only)
func (h *SystemHandler) ResumeTenant(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid tenant ID", nil)
		return
	}

	existing, err := h.tenantRepo.GetByID(c.Request.Context(), tenantID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"Tenant not found", nil)
		return
	}

	existing.Status = models.TenantStatusActive
	if err := h.tenantRepo.Update(c.Request.Context(), existing); err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to resume tenant: "+err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, existing)
}

