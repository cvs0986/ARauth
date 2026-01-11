package capability

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
)

// ValidationService provides validation logic for capability operations
type ValidationService struct {
	capabilityService ServiceInterface
}

// NewValidationService creates a new validation service
func NewValidationService(capabilityService ServiceInterface) *ValidationService {
	return &ValidationService{
		capabilityService: capabilityService,
	}
}

// ValidateTenantFeatureEnablement validates that a tenant can enable a feature
// Rules:
// 1. System must support the capability
// 2. Tenant must be allowed to use the capability
// 3. Tenant cannot exceed system limits (e.g., max_token_ttl)
func (v *ValidationService) ValidateTenantFeatureEnablement(ctx context.Context, tenantID uuid.UUID, featureKey string, config *json.RawMessage) error {
	// 1. Check if system supports it
	supported, err := v.capabilityService.IsCapabilitySupported(ctx, featureKey)
	if err != nil {
		return fmt.Errorf("failed to check system capability: %w", err)
	}
	if !supported {
		return fmt.Errorf("capability %s is not supported by the system", featureKey)
	}

	// 2. Check if tenant is allowed to use it
	allowed, err := v.capabilityService.IsCapabilityAllowedForTenant(ctx, tenantID, featureKey)
	if err != nil {
		return fmt.Errorf("failed to check tenant capability: %w", err)
	}
	if !allowed {
		return fmt.Errorf("capability %s is not allowed for this tenant", featureKey)
	}

	// 3. Validate configuration against system limits
	if config != nil {
		if err := v.validateConfiguration(ctx, tenantID, featureKey, config); err != nil {
			return fmt.Errorf("configuration validation failed: %w", err)
		}
	}

	return nil
}

// ValidateTenantCapabilityAssignment validates that a system admin can assign a capability to a tenant
// Rules:
// 1. System must support the capability
// 2. Tenant value cannot exceed system limits
func (v *ValidationService) ValidateTenantCapabilityAssignment(ctx context.Context, tenantID uuid.UUID, capabilityKey string, value *json.RawMessage) error {
	// 1. Check if system supports it
	supported, err := v.capabilityService.IsCapabilitySupported(ctx, capabilityKey)
	if err != nil {
		return fmt.Errorf("failed to check system capability: %w", err)
	}
	if !supported {
		return fmt.Errorf("capability %s is not supported by the system", capabilityKey)
	}

	// 2. Validate value against system limits
	if value != nil {
		if err := v.validateCapabilityValue(ctx, capabilityKey, value); err != nil {
			return fmt.Errorf("value validation failed: %w", err)
		}
	}

	return nil
}

// ValidateUserEnrollment validates that a user can enroll in a capability
// Rules:
// 1. System must support the capability
// 2. Tenant must be allowed to use the capability
// 3. Tenant must have enabled the feature
func (v *ValidationService) ValidateUserEnrollment(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, capabilityKey string) error {
	// Use EvaluateCapability to check all layers
	evaluation, err := v.capabilityService.EvaluateCapability(ctx, tenantID, userID, capabilityKey)
	if err != nil {
		return fmt.Errorf("failed to evaluate capability: %w", err)
	}

	if !evaluation.SystemSupported {
		return fmt.Errorf("capability %s is not supported by the system", capabilityKey)
	}

	if !evaluation.TenantAllowed {
		return fmt.Errorf("capability %s is not allowed for this tenant", capabilityKey)
	}

	if !evaluation.TenantEnabled {
		return fmt.Errorf("feature %s is not enabled by this tenant", capabilityKey)
	}

	return nil
}

// validateConfiguration validates feature configuration against system limits
func (v *ValidationService) validateConfiguration(ctx context.Context, tenantID uuid.UUID, featureKey string, config *json.RawMessage) error {
	// Get system capability to check limits
	systemCap, err := v.capabilityService.GetSystemCapability(ctx, featureKey)
	if err != nil {
		return fmt.Errorf("failed to get system capability: %w", err)
	}

	// Get tenant capability to check tenant-specific limits
	// Note: This would require a method to get tenant capability, which we'll add if needed
	// For now, we'll do basic validation

	// Example: Validate token TTL limits
	if featureKey == models.CapabilityKeyMaxTokenTTL {
		var configValue map[string]interface{}
		if err := json.Unmarshal(*config, &configValue); err != nil {
			return fmt.Errorf("invalid configuration format: %w", err)
		}

		// Check if max_token_ttl is within system limits
		if maxTTL, ok := configValue["max_token_ttl"].(string); ok {
			// Parse duration and validate
			duration, err := time.ParseDuration(maxTTL)
			if err != nil {
				return fmt.Errorf("invalid duration format: %w", err)
			}

			// Get system max TTL from default value
			if systemCap.DefaultValue != nil {
				var systemValue map[string]interface{}
				if err := json.Unmarshal(systemCap.DefaultValue, &systemValue); err == nil {
					if systemMaxTTL, ok := systemValue["max_token_ttl"].(string); ok {
						systemDuration, err := time.ParseDuration(systemMaxTTL)
						if err == nil {
							if duration > systemDuration {
								return fmt.Errorf("tenant max_token_ttl (%v) exceeds system limit (%v)", duration, systemDuration)
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// validateCapabilityValue validates a capability value against system limits
func (v *ValidationService) validateCapabilityValue(ctx context.Context, capabilityKey string, value *json.RawMessage) error {
	// Get system capability to check limits
	systemCap, err := v.capabilityService.GetSystemCapability(ctx, capabilityKey)
	if err != nil {
		return fmt.Errorf("failed to get system capability: %w", err)
	}

	// Example: Validate max_token_ttl
	if capabilityKey == models.CapabilityKeyMaxTokenTTL {
		var valueMap map[string]interface{}
		if err := json.Unmarshal(*value, &valueMap); err != nil {
			return fmt.Errorf("invalid value format: %w", err)
		}

		// Check if max_token_ttl is within system limits
		if maxTTL, ok := valueMap["max_token_ttl"].(string); ok {
			duration, err := time.ParseDuration(maxTTL)
			if err != nil {
				return fmt.Errorf("invalid duration format: %w", err)
			}

			// Get system max TTL from default value
			if systemCap.DefaultValue != nil {
				var systemValue map[string]interface{}
				if err := json.Unmarshal(systemCap.DefaultValue, &systemValue); err == nil {
					if systemMaxTTL, ok := systemValue["max_token_ttl"].(string); ok {
						systemDuration, err := time.ParseDuration(systemMaxTTL)
						if err == nil {
							if duration > systemDuration {
								return fmt.Errorf("tenant max_token_ttl (%v) exceeds system limit (%v)", duration, systemDuration)
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// ValidateScopeNamespace validates that a scope namespace is allowed for a tenant
func (v *ValidationService) ValidateScopeNamespace(ctx context.Context, tenantID uuid.UUID, namespace string) error {
	// Check if namespace is in allowed list for tenant
	// This would check tenant_capabilities.allowed_scope_namespaces
	// For now, we'll do a basic check
	capabilityKey := models.CapabilityKeyAllowedScopeNamespaces
	allowed, err := v.capabilityService.IsCapabilityAllowedForTenant(ctx, tenantID, capabilityKey)
	if err != nil {
		return fmt.Errorf("failed to check scope namespace capability: %w", err)
	}

	if !allowed {
		return fmt.Errorf("scope namespace %s is not allowed for this tenant", namespace)
	}

	// TODO: Check if namespace is in the allowed list
	// This would require parsing the tenant capability value

	return nil
}




