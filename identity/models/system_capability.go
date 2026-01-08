package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// SystemCapability represents a global system-level capability
// This implements the "System" layer of the capability model
type SystemCapability struct {
	CapabilityKey string          `json:"capability_key" db:"capability_key"`
	Enabled       bool            `json:"enabled" db:"enabled"`
	DefaultValue  json.RawMessage `json:"default_value,omitempty" db:"default_value"`
	Description   *string         `json:"description,omitempty" db:"description"`
	UpdatedBy     *uuid.UUID      `json:"updated_by,omitempty" db:"updated_by"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

// IsSupported checks if the capability is supported by the system
func (sc *SystemCapability) IsSupported() bool {
	return sc.Enabled
}

// GetDefaultValue returns the default value as a map
func (sc *SystemCapability) GetDefaultValue() (map[string]interface{}, error) {
	if len(sc.DefaultValue) == 0 {
		return make(map[string]interface{}), nil
	}
	var result map[string]interface{}
	err := json.Unmarshal(sc.DefaultValue, &result)
	return result, err
}

// SetDefaultValue sets the default value from a map
func (sc *SystemCapability) SetDefaultValue(value map[string]interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	sc.DefaultValue = data
	return nil
}

// Capability key constants
const (
	CapabilityKeyMFA                  = "mfa"
	CapabilityKeyTOTP                 = "totp"
	CapabilityKeySAML                 = "saml"
	CapabilityKeyOIDC                 = "oidc"
	CapabilityKeyOAuth2               = "oauth2"
	CapabilityKeyPasswordless         = "passwordless"
	CapabilityKeyLDAP                 = "ldap"
	CapabilityKeyMaxTokenTTL          = "max_token_ttl"
	CapabilityKeyAllowedGrantTypes    = "allowed_grant_types"
	CapabilityKeyAllowedScopeNamespaces = "allowed_scope_namespaces"
	CapabilityKeyPKCEMandatory        = "pkce_mandatory"
)

