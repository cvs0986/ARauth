package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/identity/capability"
)

// RequireCapability creates middleware that requires a specific capability to be available
// This checks the three-layer model: System → Tenant → User
func RequireCapability(capabilityService capability.ServiceInterface, capabilityKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user claims from context
		claimsObj, exists := c.Get("user_claims")
		if !exists {
			RespondWithError(c, http.StatusUnauthorized, "unauthorized", "User claims not found", nil)
			c.Abort()
			return
		}

		userClaims := claimsObj.(*claims.Claims)

		// Get tenant ID from context or claims
		var tenantID uuid.UUID
		if tenantIDFromCtx, exists := GetTenantID(c); exists {
			tenantID = tenantIDFromCtx
		} else if userClaims.TenantID != "" {
			var err error
			tenantID, err = uuid.Parse(userClaims.TenantID)
			if err != nil {
				RespondWithError(c, http.StatusBadRequest, "invalid_tenant_id", "Invalid tenant ID in token", nil)
				c.Abort()
				return
			}
		} else {
			// For SYSTEM users, we might not have a tenant context
			// In that case, we can only check system-level support
			if userClaims.PrincipalType == "SYSTEM" {
				// System users can access if capability is supported at system level
				supported, err := capabilityService.IsCapabilitySupported(c.Request.Context(), capabilityKey)
				if err != nil || !supported {
					RespondWithError(c, http.StatusForbidden, "capability_not_supported",
						"Capability '"+capabilityKey+"' is not supported at the system level", nil)
					c.Abort()
					return
				}
				c.Next()
				return
			}

			RespondWithError(c, http.StatusBadRequest, "tenant_required",
				"Tenant context is required for capability evaluation", nil)
			c.Abort()
			return
		}

		// Get user ID from claims
		userID, err := uuid.Parse(userClaims.Subject)
		if err != nil {
			RespondWithError(c, http.StatusBadRequest, "invalid_user_id", "Invalid user ID in token", nil)
			c.Abort()
			return
		}

		// Evaluate capability using the service
		evaluation, err := capabilityService.EvaluateCapability(c.Request.Context(), tenantID, userID, capabilityKey)
		if err != nil {
			RespondWithError(c, http.StatusInternalServerError, "capability_evaluation_error",
				"Failed to evaluate capability: "+err.Error(), nil)
			c.Abort()
			return
		}

		// Check if capability can be used
		if !evaluation.CanUse {
			RespondWithError(c, http.StatusForbidden, "capability_not_available",
				"Capability '"+capabilityKey+"' is not available. Reason: "+evaluation.Reason, nil)
			c.Abort()
			return
		}

		// Store evaluation result in context for use by handlers
		c.Set("capability_evaluation", evaluation)
		c.Set("capability_key", capabilityKey)

		c.Next()
	}
}

// RequireFeatureEnabled creates middleware that requires a feature to be enabled for the tenant
// This is a convenience wrapper around RequireCapability that specifically checks tenant enablement
func RequireFeatureEnabled(capabilityService capability.ServiceInterface, capabilityKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user claims from context
		claimsObj, exists := c.Get("user_claims")
		if !exists {
			RespondWithError(c, http.StatusUnauthorized, "unauthorized", "User claims not found", nil)
			c.Abort()
			return
		}

		userClaims := claimsObj.(*claims.Claims)

		// Get tenant ID
		var tenantID uuid.UUID
		if tenantIDFromCtx, exists := GetTenantID(c); exists {
			tenantID = tenantIDFromCtx
		} else if userClaims.TenantID != "" {
			var err error
			tenantID, err = uuid.Parse(userClaims.TenantID)
			if err != nil {
				RespondWithError(c, http.StatusBadRequest, "invalid_tenant_id", "Invalid tenant ID in token", nil)
				c.Abort()
				return
			}
		} else {
			RespondWithError(c, http.StatusBadRequest, "tenant_required",
				"Tenant context is required for feature enablement check", nil)
			c.Abort()
			return
		}

		// Check if feature is enabled for tenant
		enabled, err := capabilityService.IsFeatureEnabledByTenant(c.Request.Context(), tenantID, capabilityKey)
		if err != nil {
			RespondWithError(c, http.StatusInternalServerError, "feature_check_error",
				"Failed to check feature enablement: "+err.Error(), nil)
			c.Abort()
			return
		}

		if !enabled {
			RespondWithError(c, http.StatusForbidden, "feature_not_enabled",
				"Feature '"+capabilityKey+"' is not enabled for this tenant", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireUserEnrollment creates middleware that requires a user to be enrolled in a capability
func RequireUserEnrollment(capabilityService capability.ServiceInterface, capabilityKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user claims from context
		claimsObj, exists := c.Get("user_claims")
		if !exists {
			RespondWithError(c, http.StatusUnauthorized, "unauthorized", "User claims not found", nil)
			c.Abort()
			return
		}

		userClaims := claimsObj.(*claims.Claims)

		// Get user ID
		userID, err := uuid.Parse(userClaims.Subject)
		if err != nil {
			RespondWithError(c, http.StatusBadRequest, "invalid_user_id", "Invalid user ID in token", nil)
			c.Abort()
			return
		}

		// Get tenant ID
		var tenantID uuid.UUID
		if tenantIDFromCtx, exists := GetTenantID(c); exists {
			tenantID = tenantIDFromCtx
		} else if userClaims.TenantID != "" {
			var err error
			tenantID, err = uuid.Parse(userClaims.TenantID)
			if err != nil {
				RespondWithError(c, http.StatusBadRequest, "invalid_tenant_id", "Invalid tenant ID in token", nil)
				c.Abort()
				return
			}
		} else {
			RespondWithError(c, http.StatusBadRequest, "tenant_required",
				"Tenant context is required for enrollment check", nil)
			c.Abort()
			return
		}

		// Check if user is enrolled
		// Note: IsUserEnrolled doesn't require tenantID, but we should verify the capability is enabled for tenant first
		// For now, we'll use EvaluateCapability which does the full check
		evaluation, err := capabilityService.EvaluateCapability(c.Request.Context(), tenantID, userID, capabilityKey)
		if err != nil {
			RespondWithError(c, http.StatusInternalServerError, "enrollment_check_error",
				"Failed to check user enrollment: "+err.Error(), nil)
			c.Abort()
			return
		}

		if !evaluation.UserEnrolled {
			RespondWithError(c, http.StatusForbidden, "user_not_enrolled",
				"User is not enrolled in capability '"+capabilityKey+"'", nil)
			c.Abort()
			return
		}

		c.Next()
		return
	}
}

// RequireUserEnrollmentSimple creates middleware that requires a user to be enrolled in a capability (simpler version)
func RequireUserEnrollmentSimple(capabilityService capability.ServiceInterface, capabilityKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user claims from context
		claimsObj, exists := c.Get("user_claims")
		if !exists {
			RespondWithError(c, http.StatusUnauthorized, "unauthorized", "User claims not found", nil)
			c.Abort()
			return
		}

		userClaims := claimsObj.(*claims.Claims)

		// Get user ID
		userID, err := uuid.Parse(userClaims.Subject)
		if err != nil {
			RespondWithError(c, http.StatusBadRequest, "invalid_user_id", "Invalid user ID in token", nil)
			c.Abort()
			return
		}

		// Check if user is enrolled (this method doesn't require tenant context)
		enrolled, err := capabilityService.IsUserEnrolled(c.Request.Context(), userID, capabilityKey)
		if err != nil {
			RespondWithError(c, http.StatusInternalServerError, "enrollment_check_error",
				"Failed to check user enrollment: "+err.Error(), nil)
			c.Abort()
			return
		}

		if !enrolled {
			RespondWithError(c, http.StatusForbidden, "user_not_enrolled",
				"User is not enrolled in capability '"+capabilityKey+"'", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCapabilityEvaluation retrieves the capability evaluation from context
func GetCapabilityEvaluation(c *gin.Context) (*capability.CapabilityEvaluation, bool) {
	eval, exists := c.Get("capability_evaluation")
	if !exists {
		return nil, false
	}
	evaluation, ok := eval.(*capability.CapabilityEvaluation)
	return evaluation, ok
}

