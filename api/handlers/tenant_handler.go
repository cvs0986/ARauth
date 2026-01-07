package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nuage-identity/iam/api/middleware"
	"github.com/nuage-identity/iam/identity/tenant"
	"github.com/nuage-identity/iam/storage/interfaces"
)

// TenantHandler handles tenant-related HTTP requests
type TenantHandler struct {
	tenantService tenant.ServiceInterface
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(tenantService tenant.ServiceInterface) *TenantHandler {
	return &TenantHandler{tenantService: tenantService}
}

// Create handles POST /api/v1/tenants
func (h *TenantHandler) Create(c *gin.Context) {
	var req tenant.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	tenant, err := h.tenantService.Create(c.Request.Context(), &req)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "creation_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusCreated, tenant)
}

// GetByID handles GET /api/v1/tenants/:id
func (h *TenantHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid tenant ID format", nil)
		return
	}

	tenant, err := h.tenantService.GetByID(c.Request.Context(), id)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "tenant_not_found",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// GetByDomain handles GET /api/v1/tenants/domain/:domain
func (h *TenantHandler) GetByDomain(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_domain",
			"Domain parameter is required", nil)
		return
	}

	tenant, err := h.tenantService.GetByDomain(c.Request.Context(), domain)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "tenant_not_found",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// Update handles PUT /api/v1/tenants/:id
func (h *TenantHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid tenant ID format", nil)
		return
	}

	var req tenant.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	updatedTenant, err := h.tenantService.Update(c.Request.Context(), id, &req)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "update_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, updatedTenant)
}

// Delete handles DELETE /api/v1/tenants/:id
func (h *TenantHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid tenant ID format", nil)
		return
	}

	if err := h.tenantService.Delete(c.Request.Context(), id); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "delete_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tenant deleted successfully"})
}

// List handles GET /api/v1/tenants
func (h *TenantHandler) List(c *gin.Context) {
	filters := &interfaces.TenantFilters{
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

	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}

	if search := c.Query("search"); search != "" {
		filters.Search = &search
	}

	tenants, err := h.tenantService.List(c.Request.Context(), filters)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "list_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenants": tenants,
		"page":    filters.Page,
		"page_size": filters.PageSize,
	})
}

