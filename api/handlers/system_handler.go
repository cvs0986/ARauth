package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/identity/capability"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/identity/tenant"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// SystemHandler handles system-level operations (tenant management, system settings)
type SystemHandler struct {
	tenantService      tenant.ServiceInterface
	tenantRepo         interfaces.TenantRepository
	tenantSettingsRepo interfaces.TenantSettingsRepository
	capabilityService  capability.ServiceInterface
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(tenantService tenant.ServiceInterface, tenantRepo interfaces.TenantRepository, tenantSettingsRepo interfaces.TenantSettingsRepository, capabilityService capability.ServiceInterface) *SystemHandler {
	return &SystemHandler{
		tenantService:      tenantService,
		tenantRepo:         tenantRepo,
		tenantSettingsRepo: tenantSettingsRepo,
		capabilityService:  capabilityService,
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
		Name     string                 `json:"name" binding:"required"`
		Domain   string                 `json:"domain" binding:"required"`
		Status   string                 `json:"status,omitempty"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Use tenant service to create tenant (this will automatically initialize roles and permissions)
	createReq := &tenant.CreateTenantRequest{
		Name:     req.Name,
		Domain:   req.Domain,
		Metadata: req.Metadata,
	}

	createdTenant, err := h.tenantService.Create(c.Request.Context(), createReq)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to create tenant: "+err.Error(), nil)
		return
	}

	c.JSON(http.StatusCreated, createdTenant)
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

// GetTenantSettings handles GET /system/tenants/:id/settings - Get tenant settings (system admin only)
func (h *SystemHandler) GetTenantSettings(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid tenant ID", nil)
		return
	}

	// Verify tenant exists
	_, err = h.tenantRepo.GetByID(c.Request.Context(), tenantID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"Tenant not found", nil)
		return
	}

	// Get tenant settings
	settings, err := h.tenantSettingsRepo.GetByTenantID(c.Request.Context(), tenantID)
	if err != nil {
		// If settings don't exist, return default/empty settings
		// This allows SYSTEM admin to configure settings for tenants that don't have them yet
		c.JSON(http.StatusOK, gin.H{
			"tenant_id": tenantID,
			"message":   "Tenant settings not configured. Use PUT to create/update.",
		})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// GetTenantSettingsFromContext handles GET /api/v1/tenant/settings - Get tenant settings from context (TENANT users)
func (h *SystemHandler) GetTenantSettingsFromContext(c *gin.Context) {
	// Get tenant ID from context (set by TenantMiddleware)
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	// Get tenant settings
	settings, err := h.tenantSettingsRepo.GetByTenantID(c.Request.Context(), tenantID)
	if err != nil {
		// If settings don't exist, return default/empty settings
		c.JSON(http.StatusOK, gin.H{
			"tenant_id": tenantID,
			"message":   "Tenant settings not configured. Use PUT to create/update.",
		})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateTenantSettingsFromContext handles PUT /api/v1/tenant/settings - Update tenant settings from context (TENANT users)
func (h *SystemHandler) UpdateTenantSettingsFromContext(c *gin.Context) {
	// Get tenant ID from context (set by TenantMiddleware)
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	var req struct {
		AccessTokenTTLMinutes            *int  `json:"access_token_ttl_minutes,omitempty"`
		RefreshTokenTTLDays              *int  `json:"refresh_token_ttl_days,omitempty"`
		IDTokenTTLMinutes               *int  `json:"id_token_ttl_minutes,omitempty"`
		RememberMeEnabled               *bool `json:"remember_me_enabled,omitempty"`
		RememberMeRefreshTokenTTLDays   *int  `json:"remember_me_refresh_token_ttl_days,omitempty"`
		RememberMeAccessTokenTTLMinutes *int  `json:"remember_me_access_token_ttl_minutes,omitempty"`
		TokenRotationEnabled            *bool `json:"token_rotation_enabled,omitempty"`
		RequireMFAForExtendedSessions   *bool `json:"require_mfa_for_extended_sessions,omitempty"`
		// Security settings
		MinPasswordLength                *int  `json:"min_password_length,omitempty"`
		RequireUppercase                  *bool `json:"require_uppercase,omitempty"`
		RequireLowercase                  *bool `json:"require_lowercase,omitempty"`
		RequireNumbers                    *bool `json:"require_numbers,omitempty"`
		RequireSpecialChars               *bool `json:"require_special_chars,omitempty"`
		PasswordExpiryDays                *int  `json:"password_expiry_days,omitempty"`
		MFARequired                       *bool `json:"mfa_required,omitempty"`
		RateLimitRequests                 *int  `json:"rate_limit_requests,omitempty"`
		RateLimitWindowSeconds            *int  `json:"rate_limit_window_seconds,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Get existing settings or create new
	settings, err := h.tenantSettingsRepo.GetByTenantID(c.Request.Context(), tenantID)
	if err != nil {
		// Settings don't exist, create new with defaults
		settings = &interfaces.TenantSettings{
			TenantID:                          tenantID,
			AccessTokenTTLMinutes:            15,
			RefreshTokenTTLDays:              30,
			IDTokenTTLMinutes:               60,
			RememberMeEnabled:                true,
			RememberMeRefreshTokenTTLDays:    90,
			RememberMeAccessTokenTTLMinutes:  60,
			TokenRotationEnabled:             true,
			RequireMFAForExtendedSessions:    false,
			MinPasswordLength:               12,
			RequireUppercase:                 true,
			RequireLowercase:                 true,
			RequireNumbers:                   true,
			RequireSpecialChars:              true,
			PasswordExpiryDays:               nil,
			MFARequired:                     false,
			RateLimitRequests:                100,
			RateLimitWindowSeconds:           60,
		}
	}

	// Update fields if provided
	if req.AccessTokenTTLMinutes != nil {
		settings.AccessTokenTTLMinutes = *req.AccessTokenTTLMinutes
	}
	if req.RefreshTokenTTLDays != nil {
		settings.RefreshTokenTTLDays = *req.RefreshTokenTTLDays
	}
	if req.IDTokenTTLMinutes != nil {
		settings.IDTokenTTLMinutes = *req.IDTokenTTLMinutes
	}
	if req.RememberMeEnabled != nil {
		settings.RememberMeEnabled = *req.RememberMeEnabled
	}
	if req.RememberMeRefreshTokenTTLDays != nil {
		settings.RememberMeRefreshTokenTTLDays = *req.RememberMeRefreshTokenTTLDays
	}
	if req.RememberMeAccessTokenTTLMinutes != nil {
		settings.RememberMeAccessTokenTTLMinutes = *req.RememberMeAccessTokenTTLMinutes
	}
	if req.TokenRotationEnabled != nil {
		settings.TokenRotationEnabled = *req.TokenRotationEnabled
	}
	if req.RequireMFAForExtendedSessions != nil {
		settings.RequireMFAForExtendedSessions = *req.RequireMFAForExtendedSessions
	}
	if req.MinPasswordLength != nil {
		settings.MinPasswordLength = *req.MinPasswordLength
	}
	if req.RequireUppercase != nil {
		settings.RequireUppercase = *req.RequireUppercase
	}
	if req.RequireLowercase != nil {
		settings.RequireLowercase = *req.RequireLowercase
	}
	if req.RequireNumbers != nil {
		settings.RequireNumbers = *req.RequireNumbers
	}
	if req.RequireSpecialChars != nil {
		settings.RequireSpecialChars = *req.RequireSpecialChars
	}
	if req.PasswordExpiryDays != nil {
		settings.PasswordExpiryDays = req.PasswordExpiryDays
	}
	if req.MFARequired != nil {
		// Validate that MFA feature is enabled before requiring MFA for all users
		if *req.MFARequired {
			mfaEnabled, err := h.capabilityService.IsFeatureEnabledByTenant(c.Request.Context(), tenantID, models.FeatureKeyMFA)
			if err != nil {
				middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
					"Failed to check MFA feature enablement: "+err.Error(), nil)
				return
			}
			if !mfaEnabled {
				middleware.RespondWithError(c, http.StatusBadRequest, "mfa_feature_not_enabled",
					"Cannot require MFA for all users: MFA feature must be enabled for the tenant first. Please enable the MFA feature in Tenant Capabilities before setting this requirement.", nil)
				return
			}
		}
		settings.MFARequired = *req.MFARequired
	}
	if req.RateLimitRequests != nil {
		settings.RateLimitRequests = *req.RateLimitRequests
	}
	if req.RateLimitWindowSeconds != nil {
		settings.RateLimitWindowSeconds = *req.RateLimitWindowSeconds
	}

	// Save settings
	if settings.ID == uuid.Nil {
		if err := h.tenantSettingsRepo.Create(c.Request.Context(), settings); err != nil {
			middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
				"Failed to create tenant settings: "+err.Error(), nil)
			return
		}
	} else {
		if err := h.tenantSettingsRepo.Update(c.Request.Context(), settings); err != nil {
			middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
				"Failed to update tenant settings: "+err.Error(), nil)
			return
		}
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateTenantSettings handles PUT /system/tenants/:id/settings or PUT /api/v1/tenants/:id/settings
// For SYSTEM users: can update any tenant's settings
// For TENANT users: can only update their own tenant's settings
func (h *SystemHandler) UpdateTenantSettings(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid tenant ID", nil)
		return
	}

	// Verify tenant exists
	_, err = h.tenantRepo.GetByID(c.Request.Context(), tenantID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"Tenant not found", nil)
		return
	}

	// For TENANT users: verify they can only update their own tenant's settings
	// Get tenant ID from context (set by TenantMiddleware)
	if contextTenantID, exists := middleware.GetTenantID(c); exists {
		if contextTenantID != tenantID {
			middleware.RespondWithError(c, http.StatusForbidden, "access_denied",
				"You can only update your own tenant's settings", nil)
			return
		}
	}

	var req struct {
		AccessTokenTTLMinutes            *int  `json:"access_token_ttl_minutes,omitempty"`
		RefreshTokenTTLDays              *int  `json:"refresh_token_ttl_days,omitempty"`
		IDTokenTTLMinutes               *int  `json:"id_token_ttl_minutes,omitempty"`
		RememberMeEnabled               *bool `json:"remember_me_enabled,omitempty"`
		RememberMeRefreshTokenTTLDays   *int  `json:"remember_me_refresh_token_ttl_days,omitempty"`
		RememberMeAccessTokenTTLMinutes *int  `json:"remember_me_access_token_ttl_minutes,omitempty"`
		TokenRotationEnabled            *bool `json:"token_rotation_enabled,omitempty"`
		RequireMFAForExtendedSessions   *bool `json:"require_mfa_for_extended_sessions,omitempty"`
		// Security settings
		MinPasswordLength                *int  `json:"min_password_length,omitempty"`
		RequireUppercase                  *bool `json:"require_uppercase,omitempty"`
		RequireLowercase                  *bool `json:"require_lowercase,omitempty"`
		RequireNumbers                    *bool `json:"require_numbers,omitempty"`
		RequireSpecialChars               *bool `json:"require_special_chars,omitempty"`
		PasswordExpiryDays                *int  `json:"password_expiry_days,omitempty"`
		MFARequired                       *bool `json:"mfa_required,omitempty"`
		RateLimitRequests                 *int  `json:"rate_limit_requests,omitempty"`
		RateLimitWindowSeconds            *int  `json:"rate_limit_window_seconds,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Get existing settings or create new
	settings, err := h.tenantSettingsRepo.GetByTenantID(c.Request.Context(), tenantID)
	if err != nil {
		// Settings don't exist, create new with defaults
		settings = &interfaces.TenantSettings{
			TenantID:                          tenantID,
			AccessTokenTTLMinutes:            15,  // Default: 15 minutes
			RefreshTokenTTLDays:              30,  // Default: 30 days
			IDTokenTTLMinutes:               60,  // Default: 60 minutes
			RememberMeEnabled:                true,
			RememberMeRefreshTokenTTLDays:    90,  // Default: 90 days
			RememberMeAccessTokenTTLMinutes:  60,  // Default: 60 minutes
			TokenRotationEnabled:             true,
			RequireMFAForExtendedSessions:    false,
			// Security settings defaults
			MinPasswordLength:               12,
			RequireUppercase:                 true,
			RequireLowercase:                 true,
			RequireNumbers:                   true,
			RequireSpecialChars:              true,
			PasswordExpiryDays:               nil, // Never expires by default
			MFARequired:                     false,
			RateLimitRequests:                100,
			RateLimitWindowSeconds:           60,
		}
	}

	// Update fields if provided
	if req.AccessTokenTTLMinutes != nil {
		settings.AccessTokenTTLMinutes = *req.AccessTokenTTLMinutes
	}
	if req.RefreshTokenTTLDays != nil {
		settings.RefreshTokenTTLDays = *req.RefreshTokenTTLDays
	}
	if req.IDTokenTTLMinutes != nil {
		settings.IDTokenTTLMinutes = *req.IDTokenTTLMinutes
	}
	if req.RememberMeEnabled != nil {
		settings.RememberMeEnabled = *req.RememberMeEnabled
	}
	if req.RememberMeRefreshTokenTTLDays != nil {
		settings.RememberMeRefreshTokenTTLDays = *req.RememberMeRefreshTokenTTLDays
	}
	if req.RememberMeAccessTokenTTLMinutes != nil {
		settings.RememberMeAccessTokenTTLMinutes = *req.RememberMeAccessTokenTTLMinutes
	}
	if req.TokenRotationEnabled != nil {
		settings.TokenRotationEnabled = *req.TokenRotationEnabled
	}
	if req.RequireMFAForExtendedSessions != nil {
		settings.RequireMFAForExtendedSessions = *req.RequireMFAForExtendedSessions
	}
	// Update security settings if provided
	if req.MinPasswordLength != nil {
		settings.MinPasswordLength = *req.MinPasswordLength
	}
	if req.RequireUppercase != nil {
		settings.RequireUppercase = *req.RequireUppercase
	}
	if req.RequireLowercase != nil {
		settings.RequireLowercase = *req.RequireLowercase
	}
	if req.RequireNumbers != nil {
		settings.RequireNumbers = *req.RequireNumbers
	}
	if req.RequireSpecialChars != nil {
		settings.RequireSpecialChars = *req.RequireSpecialChars
	}
	if req.PasswordExpiryDays != nil {
		settings.PasswordExpiryDays = req.PasswordExpiryDays
	}
	if req.MFARequired != nil {
		// Validate that MFA feature is enabled before requiring MFA for all users
		if *req.MFARequired {
			mfaEnabled, err := h.capabilityService.IsFeatureEnabledByTenant(c.Request.Context(), tenantID, models.FeatureKeyMFA)
			if err != nil {
				middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
					"Failed to check MFA feature enablement: "+err.Error(), nil)
				return
			}
			if !mfaEnabled {
				middleware.RespondWithError(c, http.StatusBadRequest, "mfa_feature_not_enabled",
					"Cannot require MFA for all users: MFA feature must be enabled for the tenant first. Please enable the MFA feature in Tenant Capabilities before setting this requirement.", nil)
				return
			}
		}
		settings.MFARequired = *req.MFARequired
	}
	if req.RateLimitRequests != nil {
		settings.RateLimitRequests = *req.RateLimitRequests
	}
	if req.RateLimitWindowSeconds != nil {
		settings.RateLimitWindowSeconds = *req.RateLimitWindowSeconds
	}

	// Save settings
	if settings.ID == uuid.Nil {
		// Create new settings
		if err := h.tenantSettingsRepo.Create(c.Request.Context(), settings); err != nil {
			middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
				"Failed to create tenant settings: "+err.Error(), nil)
			return
		}
	} else {
		// Update existing settings
		if err := h.tenantSettingsRepo.Update(c.Request.Context(), settings); err != nil {
			middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
				"Failed to update tenant settings: "+err.Error(), nil)
			return
		}
	}

	c.JSON(http.StatusOK, settings)
}


