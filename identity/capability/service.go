package capability

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// Service provides capability evaluation business logic
// Implements the three-layer capability model: System → System→Tenant → Tenant → User
type Service struct {
	systemCapabilityRepo        interfaces.SystemCapabilityRepository
	tenantCapabilityRepo        interfaces.TenantCapabilityRepository
	tenantFeatureEnablementRepo interfaces.TenantFeatureEnablementRepository
	userCapabilityStateRepo     interfaces.UserCapabilityStateRepository
}

// NewService creates a new capability service
func NewService(
	systemCapabilityRepo interfaces.SystemCapabilityRepository,
	tenantCapabilityRepo interfaces.TenantCapabilityRepository,
	tenantFeatureEnablementRepo interfaces.TenantFeatureEnablementRepository,
	userCapabilityStateRepo interfaces.UserCapabilityStateRepository,
) *Service {
	return &Service{
		systemCapabilityRepo:        systemCapabilityRepo,
		tenantCapabilityRepo:        tenantCapabilityRepo,
		tenantFeatureEnablementRepo: tenantFeatureEnablementRepo,
		userCapabilityStateRepo:     userCapabilityStateRepo,
	}
}

// ============================================================================
// System Level Methods
// ============================================================================

// IsCapabilitySupported checks if a capability is supported at the system level
func (s *Service) IsCapabilitySupported(ctx context.Context, capabilityKey string) (bool, error) {
	capability, err := s.systemCapabilityRepo.GetByKey(ctx, capabilityKey)
	if err != nil {
		return false, fmt.Errorf("failed to check system capability: %w", err)
	}
	return capability.IsSupported(), nil
}

// GetSystemCapability retrieves a system capability by key
func (s *Service) GetSystemCapability(ctx context.Context, capabilityKey string) (*models.SystemCapability, error) {
	capability, err := s.systemCapabilityRepo.GetByKey(ctx, capabilityKey)
	if err != nil {
		return nil, fmt.Errorf("system capability not found: %w", err)
	}
	return capability, nil
}

// GetAllSystemCapabilities retrieves all system capabilities
func (s *Service) GetAllSystemCapabilities(ctx context.Context) ([]*models.SystemCapability, error) {
	capabilities, err := s.systemCapabilityRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system capabilities: %w", err)
	}
	return capabilities, nil
}

// UpdateSystemCapability updates a system capability
func (s *Service) UpdateSystemCapability(ctx context.Context, capability *models.SystemCapability) error {
	capability.UpdatedAt = time.Now()
	if err := s.systemCapabilityRepo.Update(ctx, capability); err != nil {
		return fmt.Errorf("failed to update system capability: %w", err)
	}
	return nil
}

// ============================================================================
// System → Tenant Level Methods
// ============================================================================

// IsCapabilityAllowedForTenant checks if a capability is allowed for a tenant
func (s *Service) IsCapabilityAllowedForTenant(ctx context.Context, tenantID uuid.UUID, capabilityKey string) (bool, error) {
	// First check if system supports it
	supported, err := s.IsCapabilitySupported(ctx, capabilityKey)
	if err != nil {
		return false, err
	}
	if !supported {
		return false, nil
	}

	// Then check if tenant has it assigned
	capability, err := s.tenantCapabilityRepo.GetByTenantIDAndKey(ctx, tenantID, capabilityKey)
	if err != nil {
		// If not found, it's not allowed
		return false, nil
	}
	return capability.IsAllowed(), nil
}

// GetAllowedCapabilitiesForTenant retrieves all allowed capabilities for a tenant
func (s *Service) GetAllowedCapabilitiesForTenant(ctx context.Context, tenantID uuid.UUID) (map[string]bool, error) {
	capabilities, err := s.tenantCapabilityRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant capabilities: %w", err)
	}

	result := make(map[string]bool)
	for _, cap := range capabilities {
		result[cap.CapabilityKey] = cap.IsAllowed()
	}
	return result, nil
}

// GetTenantCapabilities retrieves all tenant capability objects for a tenant
func (s *Service) GetTenantCapabilities(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantCapability, error) {
	return s.tenantCapabilityRepo.GetByTenantID(ctx, tenantID)
}

// SetTenantCapability assigns a capability to a tenant
func (s *Service) SetTenantCapability(ctx context.Context, tenantID uuid.UUID, capabilityKey string, enabled bool, value *json.RawMessage, configuredBy uuid.UUID) error {
	// First verify system supports it
	supported, err := s.IsCapabilitySupported(ctx, capabilityKey)
	if err != nil {
		return err
	}
	if !supported {
		return fmt.Errorf("capability %s is not supported by the system", capabilityKey)
	}

	// Check if already exists
	existing, err := s.tenantCapabilityRepo.GetByTenantIDAndKey(ctx, tenantID, capabilityKey)
	if err == nil {
		// Update existing
		existing.Enabled = enabled
		if value != nil {
			existing.Value = *value
		}
		existing.ConfiguredBy = &configuredBy
		existing.ConfiguredAt = time.Now()
		return s.tenantCapabilityRepo.Update(ctx, existing)
	}

	// Create new
	capability := &models.TenantCapability{
		TenantID:      tenantID,
		CapabilityKey: capabilityKey,
		Enabled:       enabled,
		ConfiguredBy:  &configuredBy,
		ConfiguredAt:  time.Now(),
	}
	if value != nil {
		capability.Value = *value
	}

	return s.tenantCapabilityRepo.Create(ctx, capability)
}

// DeleteTenantCapability removes a capability from a tenant
func (s *Service) DeleteTenantCapability(ctx context.Context, tenantID uuid.UUID, capabilityKey string) error {
	return s.tenantCapabilityRepo.Delete(ctx, tenantID, capabilityKey)
}

// ============================================================================
// Tenant Level Methods
// ============================================================================

// IsFeatureEnabledByTenant checks if a feature is enabled by a tenant
func (s *Service) IsFeatureEnabledByTenant(ctx context.Context, tenantID uuid.UUID, featureKey string) (bool, error) {
	// First check if tenant is allowed to use this capability
	allowed, err := s.IsCapabilityAllowedForTenant(ctx, tenantID, featureKey)
	if err != nil {
		return false, err
	}
	if !allowed {
		return false, nil
	}

	// Then check if tenant has enabled it
	enablement, err := s.tenantFeatureEnablementRepo.GetByTenantIDAndKey(ctx, tenantID, featureKey)
	if err != nil {
		// If not found, it's not enabled
		return false, nil
	}
	return enablement.IsEnabled(), nil
}

// GetEnabledFeaturesForTenant retrieves all enabled features for a tenant
func (s *Service) GetEnabledFeaturesForTenant(ctx context.Context, tenantID uuid.UUID) (map[string]bool, error) {
	enablements, err := s.tenantFeatureEnablementRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant feature enablements: %w", err)
	}

	result := make(map[string]bool)
	for _, enablement := range enablements {
		result[enablement.FeatureKey] = enablement.IsEnabled()
	}
	return result, nil
}

// GetTenantFeatureEnablements retrieves all feature enablements for a tenant
func (s *Service) GetTenantFeatureEnablements(ctx context.Context, tenantID uuid.UUID) ([]*models.TenantFeatureEnablement, error) {
	return s.tenantFeatureEnablementRepo.GetByTenantID(ctx, tenantID)
}

// EnableFeatureForTenant enables a feature for a tenant
func (s *Service) EnableFeatureForTenant(ctx context.Context, tenantID uuid.UUID, featureKey string, config *json.RawMessage, enabledBy uuid.UUID) error {
	// First verify tenant is allowed to use this capability
	allowed, err := s.IsCapabilityAllowedForTenant(ctx, tenantID, featureKey)
	if err != nil {
		return err
	}
	if !allowed {
		return fmt.Errorf("capability %s is not allowed for this tenant", featureKey)
	}

	// Check if already exists
	existing, err := s.tenantFeatureEnablementRepo.GetByTenantIDAndKey(ctx, tenantID, featureKey)
	if err == nil {
		// Update existing
		existing.Enabled = true
		if config != nil {
			existing.Configuration = *config
		}
		existing.EnabledBy = &enabledBy
		existing.EnabledAt = time.Now()
		return s.tenantFeatureEnablementRepo.Update(ctx, existing)
	}

	// Create new
	enablement := &models.TenantFeatureEnablement{
		TenantID:     tenantID,
		FeatureKey:   featureKey,
		Enabled:      true,
		EnabledBy:    &enabledBy,
		EnabledAt:    time.Now(),
	}
	if config != nil {
		enablement.Configuration = *config
	}

	return s.tenantFeatureEnablementRepo.Create(ctx, enablement)
}

// DisableFeatureForTenant disables a feature for a tenant
func (s *Service) DisableFeatureForTenant(ctx context.Context, tenantID uuid.UUID, featureKey string) error {
	enablement, err := s.tenantFeatureEnablementRepo.GetByTenantIDAndKey(ctx, tenantID, featureKey)
	if err != nil {
		return fmt.Errorf("feature enablement not found: %w", err)
	}

	enablement.Enabled = false
	enablement.EnabledAt = time.Now()
	return s.tenantFeatureEnablementRepo.Update(ctx, enablement)
}

// ============================================================================
// User Level Methods
// ============================================================================

// IsUserEnrolled checks if a user is enrolled in a capability
func (s *Service) IsUserEnrolled(ctx context.Context, userID uuid.UUID, capabilityKey string) (bool, error) {
	state, err := s.userCapabilityStateRepo.GetByUserIDAndKey(ctx, userID, capabilityKey)
	if err != nil {
		return false, nil // Not enrolled if not found
	}
	return state.IsEnrolled(), nil
}

// GetUserCapabilityState retrieves a user's capability state
func (s *Service) GetUserCapabilityState(ctx context.Context, userID uuid.UUID, capabilityKey string) (*models.UserCapabilityState, error) {
	state, err := s.userCapabilityStateRepo.GetByUserIDAndKey(ctx, userID, capabilityKey)
	if err != nil {
		return nil, fmt.Errorf("user capability state not found: %w", err)
	}
	return state, nil
}

// GetUserCapabilityStates retrieves all capability states for a user
func (s *Service) GetUserCapabilityStates(ctx context.Context, userID uuid.UUID) ([]*models.UserCapabilityState, error) {
	states, err := s.userCapabilityStateRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user capability states: %w", err)
	}
	return states, nil
}

// EnrollUserInCapability enrolls a user in a capability
func (s *Service) EnrollUserInCapability(ctx context.Context, userID uuid.UUID, capabilityKey string, stateData *json.RawMessage) error {
	now := time.Now()
	state := &models.UserCapabilityState{
		UserID:       userID,
		CapabilityKey: capabilityKey,
		Enrolled:     true,
		EnrolledAt:   &now,
	}
	if stateData != nil {
		state.StateData = *stateData
	}

	// Check if already exists
	existing, err := s.userCapabilityStateRepo.GetByUserIDAndKey(ctx, userID, capabilityKey)
	if err == nil {
		// Update existing
		existing.Enrolled = true
		if stateData != nil {
			existing.StateData = *stateData
		}
		existing.EnrolledAt = &now
		return s.userCapabilityStateRepo.Update(ctx, existing)
	}

	return s.userCapabilityStateRepo.Create(ctx, state)
}

// UnenrollUserFromCapability unenrolls a user from a capability
func (s *Service) UnenrollUserFromCapability(ctx context.Context, userID uuid.UUID, capabilityKey string) error {
	state, err := s.userCapabilityStateRepo.GetByUserIDAndKey(ctx, userID, capabilityKey)
	if err != nil {
		return fmt.Errorf("user capability state not found: %w", err)
	}

	state.Enrolled = false
	state.EnrolledAt = nil
	return s.userCapabilityStateRepo.Update(ctx, state)
}

// ============================================================================
// Evaluation Methods (Combines All Levels)
// ============================================================================

// EvaluateCapability evaluates a capability across all layers (System → Tenant → User)
// This is the main method that combines all levels to determine if a user can use a capability
func (s *Service) EvaluateCapability(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, capabilityKey string) (*CapabilityEvaluation, error) {
	eval := &CapabilityEvaluation{
		CapabilityKey: capabilityKey,
	}

	// 1. System Level: Is it supported?
	systemCap, err := s.systemCapabilityRepo.GetByKey(ctx, capabilityKey)
	if err != nil {
		eval.SystemSupported = false
		eval.CanUse = false
		eval.Reason = fmt.Sprintf("capability %s is not supported by the system", capabilityKey)
		return eval, nil
	}
	eval.SystemSupported = systemCap.IsSupported()
	if systemCap.DefaultValue != nil {
		eval.SystemValue = systemCap.DefaultValue
	}

	if !eval.SystemSupported {
		eval.CanUse = false
		eval.Reason = fmt.Sprintf("capability %s is not supported by the system", capabilityKey)
		return eval, nil
	}

	// 2. System → Tenant Level: Is it allowed for tenant?
	tenantCap, err := s.tenantCapabilityRepo.GetByTenantIDAndKey(ctx, tenantID, capabilityKey)
	if err != nil {
		eval.TenantAllowed = false
		eval.CanUse = false
		eval.Reason = fmt.Sprintf("capability %s is not allowed for this tenant", capabilityKey)
		return eval, nil
	}
	eval.TenantAllowed = tenantCap.IsAllowed()
	if tenantCap.Value != nil {
		eval.TenantValue = tenantCap.Value
	}

	if !eval.TenantAllowed {
		eval.CanUse = false
		eval.Reason = fmt.Sprintf("capability %s is not allowed for this tenant", capabilityKey)
		return eval, nil
	}

	// 3. Tenant Level: Is it enabled by tenant?
	tenantFeature, err := s.tenantFeatureEnablementRepo.GetByTenantIDAndKey(ctx, tenantID, capabilityKey)
	if err != nil {
		eval.TenantEnabled = false
		eval.CanUse = false
		eval.Reason = fmt.Sprintf("feature %s is not enabled by this tenant", capabilityKey)
		return eval, nil
	}
	eval.TenantEnabled = tenantFeature.IsEnabled()
	if tenantFeature.Configuration != nil {
		eval.TenantConfiguration = tenantFeature.Configuration
	}

	if !eval.TenantEnabled {
		eval.CanUse = false
		eval.Reason = fmt.Sprintf("feature %s is not enabled by this tenant", capabilityKey)
		return eval, nil
	}

	// 4. User Level: Is user enrolled? (Only for capabilities that require enrollment)
	// Some capabilities (like OAuth2/OIDC) don't require user enrollment
	requiresEnrollment := requiresUserEnrollment(capabilityKey)
	if requiresEnrollment {
		userState, err := s.userCapabilityStateRepo.GetByUserIDAndKey(ctx, userID, capabilityKey)
		if err != nil {
			eval.UserEnrolled = false
			eval.CanUse = false
			eval.Reason = fmt.Sprintf("user is not enrolled in capability %s", capabilityKey)
			return eval, nil
		}
		eval.UserEnrolled = userState.IsEnrolled()
		if userState.StateData != nil {
			eval.UserStateData = userState.StateData
		}

		if !eval.UserEnrolled {
			eval.CanUse = false
			eval.Reason = fmt.Sprintf("user is not enrolled in capability %s", capabilityKey)
			return eval, nil
		}
	} else {
		// Capabilities that don't require enrollment (like OAuth2/OIDC)
		eval.UserEnrolled = true // Not applicable
	}

	// All checks passed
	eval.CanUse = true
	return eval, nil
}

// requiresUserEnrollment determines if a capability requires user enrollment
func requiresUserEnrollment(capabilityKey string) bool {
	// Capabilities that require user enrollment
	requiresEnrollment := []string{
		models.CapabilityKeyTOTP,
		models.CapabilityKeyMFA,
		models.CapabilityKeyPasswordless,
	}
	for _, key := range requiresEnrollment {
		if capabilityKey == key {
			return true
		}
	}
	// OAuth2/OIDC/SAML don't require enrollment - user just authenticates
	return false
}

