package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TenantCapability represents a capability assigned to a tenant
// This implements the "System → Tenant" layer of the capability model
type TenantCapability struct {
	TenantID      uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	CapabilityKey string          `json:"capability_key" db:"capability_key"`
	Enabled       bool            `json:"enabled" db:"enabled"`
	Value         json.RawMessage `json:"value,omitempty" db:"value"`
	ConfiguredBy  *uuid.UUID      `json:"configured_by,omitempty" db:"configured_by"`
	ConfiguredAt  time.Time       `json:"configured_at" db:"configured_at"`
}

// IsAllowed checks if the capability is allowed for the tenant
func (tc *TenantCapability) IsAllowed() bool {
	return tc.Enabled
}

// GetValue returns the capability value as a map
func (tc *TenantCapability) GetValue() (map[string]interface{}, error) {
	if len(tc.Value) == 0 {
		return make(map[string]interface{}), nil
	}
	var result map[string]interface{}
	err := json.Unmarshal(tc.Value, &result)
	return result, err
}

// SetValue sets the capability value from a map
func (tc *TenantCapability) SetValue(value map[string]interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	tc.Value = data
	return nil
}

// GetStringValue returns a string value from the capability value
func (tc *TenantCapability) GetStringValue(key string) (string, error) {
	valueMap, err := tc.GetValue()
	if err != nil {
		return "", err
	}
	if val, ok := valueMap[key].(string); ok {
		return val, nil
	}
	return "", nil
}

// GetArrayValue returns an array value from the capability value
func (tc *TenantCapability) GetArrayValue(key string) ([]string, error) {
	valueMap, err := tc.GetValue()
	if err != nil {
		return nil, err
	}
	if val, ok := valueMap[key].([]interface{}); ok {
		result := make([]string, 0, len(val))
		for _, v := range val {
			if str, ok := v.(string); ok {
				result = append(result, str)
			}
		}
		return result, nil
	}
	return nil, nil
}

