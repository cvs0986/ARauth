package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/identity/capability"
)

// CapabilityHandler handles capability management operations
type CapabilityHandler struct {
	capabilityService capability.ServiceInterface
}

// NewCapabilityHandler creates a new capability handler
func NewCapabilityHandler(capabilityService capability.ServiceInterface) *CapabilityHandler {
	return &CapabilityHandler{
		capabilityService: capabilityService,
	}
}

// ============================================================================
// System Capability Management (SYSTEM users only)
// ============================================================================

// ListSystemCapabilities handles GET /system/capabilities
// Lists all system capabilities
func (h *CapabilityHandler) ListSystemCapabilities(c *gin.Context) {
	capabilities, err := h.capabilityService.GetAllSystemCapabilities(c.Request.Context())
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to list system capabilities", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"capabilities": capabilities,
	})
}

// GetSystemCapability handles GET /system/capabilities/:key
// Gets a specific system capability
func (h *CapabilityHandler) GetSystemCapability(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_key",
			"Capability key is required", nil)
		return
	}

	cap, err := h.capabilityService.GetSystemCapability(c.Request.Context(), key)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"System capability not found", nil)
		return
	}

	c.JSON(http.StatusOK, cap)
}

// UpdateSystemCapability handles PUT /system/capabilities/:key
// Updates a system capability (system_owner only)
func (h *CapabilityHandler) UpdateSystemCapability(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_key",
			"Capability key is required", nil)
		return
	}

	var req struct {
		Enabled      *bool            `json:"enabled,omitempty"`
		DefaultValue *json.RawMessage `json:"default_value,omitempty"`
		Description  *string          `json:"description,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Get existing capability
	cap, err := h.capabilityService.GetSystemCapability(c.Request.Context(), key)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"System capability not found", nil)
		return
	}

	// Update fields if provided
	if req.Enabled != nil {
		cap.Enabled = *req.Enabled
	}
	if req.DefaultValue != nil {
		cap.DefaultValue = *req.DefaultValue
	}
	if req.Description != nil {
		cap.Description = req.Description
	}

	// Get current user ID for updated_by
	if userIDStr, ok := c.Get("user_id"); ok {
		if userIDStrStr, ok := userIDStr.(string); ok {
			if userID, err := uuid.Parse(userIDStrStr); err == nil {
				cap.UpdatedBy = &userID
			}
		}
	}

	if err := h.capabilityService.UpdateSystemCapability(c.Request.Context(), cap); err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to update system capability", nil)
		return
	}

	c.JSON(http.StatusOK, cap)
}

// ============================================================================
// Tenant Capability Assignment (SYSTEM users only)
// ============================================================================

// GetTenantCapabilities handles GET /system/tenants/:id/capabilities
// Gets all allowed capabilities for a tenant
func (h *CapabilityHandler) GetTenantCapabilities(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid tenant ID", nil)
		return
	}

	capabilities, err := h.capabilityService.GetAllowedCapabilitiesForTenant(c.Request.Context(), tenantID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to get tenant capabilities", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenant_id":   tenantID,
		"capabilities": capabilities,
	})
}

// SetTenantCapability handles PUT /system/tenants/:id/capabilities/:key
// Assigns a capability to a tenant
func (h *CapabilityHandler) SetTenantCapability(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid tenant ID", nil)
		return
	}

	capabilityKey := c.Param("key")
	if capabilityKey == "" {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_key",
			"Capability key is required", nil)
		return
	}

	var req struct {
		Enabled *bool            `json:"enabled"`
		Value   *json.RawMessage `json:"value,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	if req.Enabled == nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"enabled field is required", nil)
		return
	}

	// Get current user ID for configured_by
	userIDStr, ok := c.Get("user_id")
	if !ok {
		middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
			"User ID not found in context", nil)
		return
	}
	userIDStrVal, ok := userIDStr.(string)
	if !ok {
		middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
			"Invalid user ID in context", nil)
		return
	}
	userID, err := uuid.Parse(userIDStrVal)
	if err != nil {
		middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
			"Invalid user ID format", nil)
		return
	}

	if err := h.capabilityService.SetTenantCapability(c.Request.Context(), tenantID, capabilityKey, *req.Enabled, req.Value, userID); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			err.Error(), nil)
		return
	}

	// Return the updated capability
	cap, err := h.capabilityService.GetSystemCapability(c.Request.Context(), capabilityKey)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"tenant_id":     tenantID,
			"capability_key": capabilityKey,
			"enabled":        *req.Enabled,
			"value":          req.Value,
			"system_capability": cap,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"tenant_id":     tenantID,
			"capability_key": capabilityKey,
			"enabled":        *req.Enabled,
			"value":          req.Value,
		})
	}
}

// DeleteTenantCapability handles DELETE /system/tenants/:id/capabilities/:key
// Revokes a capability from a tenant
func (h *CapabilityHandler) DeleteTenantCapability(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid tenant ID", nil)
		return
	}

	capabilityKey := c.Param("key")
	if capabilityKey == "" {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_key",
			"Capability key is required", nil)
		return
	}

	if err := h.capabilityService.DeleteTenantCapability(c.Request.Context(), tenantID, capabilityKey); err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to delete tenant capability", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Capability revoked from tenant",
		"tenant_id": tenantID,
		"capability_key": capabilityKey,
	})
}

// EvaluateTenantCapabilities handles GET /system/tenants/:id/capabilities/evaluation
// Evaluates all capabilities for a tenant
func (h *CapabilityHandler) EvaluateTenantCapabilities(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid tenant ID", nil)
		return
	}

	// Get all system capabilities
	systemCaps, err := h.capabilityService.GetAllSystemCapabilities(c.Request.Context())
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to get system capabilities", nil)
		return
	}

	// Evaluate each capability
	evaluations := make([]*capability.CapabilityEvaluation, 0)
	for _, sysCap := range systemCaps {
		// For evaluation, we need a user ID - use a placeholder or get from query
		userIDStr := c.Query("user_id")
		var userID uuid.UUID
		if userIDStr != "" {
			userID, _ = uuid.Parse(userIDStr)
		}

		eval, err := h.capabilityService.EvaluateCapability(c.Request.Context(), tenantID, userID, sysCap.CapabilityKey)
		if err == nil {
			evaluations = append(evaluations, eval)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tenant_id":   tenantID,
		"evaluations": evaluations,
	})
}

// ============================================================================
// Tenant Feature Enablement (TENANT users)
// ============================================================================

// GetTenantFeatures handles GET /api/v1/tenant/features
// Gets all enabled features for the current tenant
func (h *CapabilityHandler) GetTenantFeatures(c *gin.Context) {
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	features, err := h.capabilityService.GetEnabledFeaturesForTenant(c.Request.Context(), tenantID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to get tenant features", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenant_id": tenantID,
		"features":  features,
	})
}

// EnableTenantFeature handles PUT /api/v1/tenant/features/:key
// Enables a feature for the current tenant
func (h *CapabilityHandler) EnableTenantFeature(c *gin.Context) {
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	featureKey := c.Param("key")
	if featureKey == "" {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_key",
			"Feature key is required", nil)
		return
	}

	var req struct {
		Configuration *json.RawMessage `json:"configuration,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Get current user ID for enabled_by
	userIDStr, ok := c.Get("user_id")
	if !ok {
		middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
			"User ID not found in context", nil)
		return
	}
	userIDStrVal, ok := userIDStr.(string)
	if !ok {
		middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
			"Invalid user ID in context", nil)
		return
	}
	userID, err := uuid.Parse(userIDStrVal)
	if err != nil {
		middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
			"Invalid user ID format", nil)
		return
	}

	if err := h.capabilityService.EnableFeatureForTenant(c.Request.Context(), tenantID, featureKey, req.Configuration, userID); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Feature enabled",
		"tenant_id":   tenantID,
		"feature_key": featureKey,
	})
}

// DisableTenantFeature handles DELETE /api/v1/tenant/features/:key
// Disables a feature for the current tenant
func (h *CapabilityHandler) DisableTenantFeature(c *gin.Context) {
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	featureKey := c.Param("key")
	if featureKey == "" {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_key",
			"Feature key is required", nil)
		return
	}

	if err := h.capabilityService.DisableFeatureForTenant(c.Request.Context(), tenantID, featureKey); err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to disable feature", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Feature disabled",
		"tenant_id":   tenantID,
		"feature_key": featureKey,
	})
}

// ============================================================================
// User Capability State (TENANT users)
// ============================================================================

// GetUserCapabilities handles GET /api/v1/users/:id/capabilities
// Gets all capability states for a user
func (h *CapabilityHandler) GetUserCapabilities(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid user ID", nil)
		return
	}

	// Authorization: User can view own capabilities, admin can view others
	var currentUserID uuid.UUID
	if userIDStr, ok := c.Get("user_id"); ok {
		if userIDStrVal, ok := userIDStr.(string); ok {
			currentUserID, _ = uuid.Parse(userIDStrVal)
		}
	}
	tenantID, _ := middleware.RequireTenant(c)
	
	// Check if user is viewing own capabilities or is admin
	if currentUserID != userID {
		// TODO: Check if current user has admin permission
		// For now, allow if in same tenant
	}

	states, err := h.capabilityService.GetUserCapabilityStates(c.Request.Context(), userID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to get user capability states", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"tenant_id": tenantID,
		"states":   states,
	})
}

// GetUserCapability handles GET /api/v1/users/:id/capabilities/:key
// Gets a specific capability state for a user
func (h *CapabilityHandler) GetUserCapability(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid user ID", nil)
		return
	}

	capabilityKey := c.Param("key")
	if capabilityKey == "" {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_key",
			"Capability key is required", nil)
		return
	}

	state, err := h.capabilityService.GetUserCapabilityState(c.Request.Context(), userID, capabilityKey)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"User capability state not found", nil)
		return
	}

	c.JSON(http.StatusOK, state)
}

// EnrollUserCapability handles POST /api/v1/users/:id/capabilities/:key/enroll
// Enrolls a user in a capability
func (h *CapabilityHandler) EnrollUserCapability(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid user ID", nil)
		return
	}

	capabilityKey := c.Param("key")
	if capabilityKey == "" {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_key",
			"Capability key is required", nil)
		return
	}

	var req struct {
		StateData *json.RawMessage `json:"state_data,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Get tenant ID for validation
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	// Validate capability is enabled for tenant before allowing enrollment
	eval, err := h.capabilityService.EvaluateCapability(c.Request.Context(), tenantID, userID, capabilityKey)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to evaluate capability", nil)
		return
	}
	if !eval.CanUse {
		middleware.RespondWithError(c, http.StatusBadRequest, "capability_not_available",
			eval.Reason, nil)
		return
	}

	if err := h.capabilityService.EnrollUserInCapability(c.Request.Context(), userID, capabilityKey, req.StateData); err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to enroll user in capability", nil)
		return
	}

	// Return updated state
	state, _ := h.capabilityService.GetUserCapabilityState(c.Request.Context(), userID, capabilityKey)
	c.JSON(http.StatusOK, gin.H{
		"message": "User enrolled in capability",
		"state":    state,
	})
}

// UnenrollUserCapability handles DELETE /api/v1/users/:id/capabilities/:key
// Unenrolls a user from a capability
func (h *CapabilityHandler) UnenrollUserCapability(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid user ID", nil)
		return
	}

	capabilityKey := c.Param("key")
	if capabilityKey == "" {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_key",
			"Capability key is required", nil)
		return
	}

	if err := h.capabilityService.UnenrollUserFromCapability(c.Request.Context(), userID, capabilityKey); err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to unenroll user from capability", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "User unenrolled from capability",
		"user_id":        userID,
		"capability_key": capabilityKey,
	})
}

