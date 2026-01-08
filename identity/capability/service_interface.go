package capability

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
)

// ServiceInterface defines the interface for capability service operations
type ServiceInterface interface {
	// System level
	IsCapabilitySupported(ctx context.Context, capabilityKey string) (bool, error)
	GetSystemCapability(ctx context.Context, capabilityKey string) (*models.SystemCapability, error)
	GetAllSystemCapabilities(ctx context.Context) ([]*models.SystemCapability, error)
	UpdateSystemCapability(ctx context.Context, capability *models.SystemCapability) error

	// System → Tenant level
	IsCapabilityAllowedForTenant(ctx context.Context, tenantID uuid.UUID, capabilityKey string) (bool, error)
	GetAllowedCapabilitiesForTenant(ctx context.Context, tenantID uuid.UUID) (map[string]bool, error)
	SetTenantCapability(ctx context.Context, tenantID uuid.UUID, capabilityKey string, enabled bool, value *json.RawMessage, configuredBy uuid.UUID) error
	DeleteTenantCapability(ctx context.Context, tenantID uuid.UUID, capabilityKey string) error

	// Tenant level
	IsFeatureEnabledByTenant(ctx context.Context, tenantID uuid.UUID, featureKey string) (bool, error)
	GetEnabledFeaturesForTenant(ctx context.Context, tenantID uuid.UUID) (map[string]bool, error)
	EnableFeatureForTenant(ctx context.Context, tenantID uuid.UUID, featureKey string, config *json.RawMessage, enabledBy uuid.UUID) error
	DisableFeatureForTenant(ctx context.Context, tenantID uuid.UUID, featureKey string) error

	// User level
	IsUserEnrolled(ctx context.Context, userID uuid.UUID, capabilityKey string) (bool, error)
	GetUserCapabilityState(ctx context.Context, userID uuid.UUID, capabilityKey string) (*models.UserCapabilityState, error)
	GetUserCapabilityStates(ctx context.Context, userID uuid.UUID) ([]*models.UserCapabilityState, error)
	EnrollUserInCapability(ctx context.Context, userID uuid.UUID, capabilityKey string, stateData *json.RawMessage) error
	UnenrollUserFromCapability(ctx context.Context, userID uuid.UUID, capabilityKey string) error

	// Evaluation (combines all levels)
	EvaluateCapability(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, capabilityKey string) (*CapabilityEvaluation, error)
}

// CapabilityEvaluation represents the result of evaluating a capability across all layers
type CapabilityEvaluation struct {
	CapabilityKey        string `json:"capability_key"`
	SystemSupported     bool   `json:"system_supported"`      // System level: Is it supported?
	TenantAllowed       bool   `json:"tenant_allowed"`        // System→Tenant level: Is it allowed for tenant?
	TenantEnabled       bool   `json:"tenant_enabled"`        // Tenant level: Is it enabled by tenant?
	UserEnrolled        bool   `json:"user_enrolled"`          // User level: Is user enrolled?
	CanUse              bool   `json:"can_use"`                // Final result: Can the user use this capability?
	Reason              string `json:"reason,omitempty"`       // Reason if can_use is false
	SystemValue         json.RawMessage `json:"system_value,omitempty"`         // System default value
	TenantValue         json.RawMessage `json:"tenant_value,omitempty"`          // Tenant-specific value
	TenantConfiguration json.RawMessage `json:"tenant_configuration,omitempty"` // Tenant feature configuration
	UserStateData       json.RawMessage `json:"user_state_data,omitempty"`       // User enrollment state
}

