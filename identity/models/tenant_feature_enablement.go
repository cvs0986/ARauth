package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TenantFeatureEnablement represents a feature enabled by a tenant
// This implements the "Tenant" layer of the capability model
type TenantFeatureEnablement struct {
	TenantID     uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	FeatureKey   string          `json:"feature_key" db:"feature_key"`
	Enabled      bool            `json:"enabled" db:"enabled"`
	Configuration json.RawMessage `json:"configuration,omitempty" db:"configuration"`
	EnabledBy    *uuid.UUID      `json:"enabled_by,omitempty" db:"enabled_by"`
	EnabledAt    time.Time       `json:"enabled_at" db:"enabled_at"`
}

// IsEnabled checks if the feature is enabled by the tenant
func (tfe *TenantFeatureEnablement) IsEnabled() bool {
	return tfe.Enabled
}

// GetConfiguration returns the configuration as a map
func (tfe *TenantFeatureEnablement) GetConfiguration() (map[string]interface{}, error) {
	if len(tfe.Configuration) == 0 {
		return make(map[string]interface{}), nil
	}
	var result map[string]interface{}
	err := json.Unmarshal(tfe.Configuration, &result)
	return result, err
}

// SetConfiguration sets the configuration from a map
func (tfe *TenantFeatureEnablement) SetConfiguration(config map[string]interface{}) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	tfe.Configuration = data
	return nil
}

// Feature key constants (same as capability keys for features)
const (
	FeatureKeyMFA          = "mfa"
	FeatureKeyTOTP         = "totp"
	FeatureKeySAML         = "saml"
	FeatureKeyOIDC         = "oidc"
	FeatureKeyOAuth2       = "oauth2"
	FeatureKeyPasswordless = "passwordless"
	FeatureKeyLDAP         = "ldap"
)

