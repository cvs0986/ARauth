package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// UserCapabilityState represents a user's enrollment state for a capability
// This implements the "User" layer of the capability model
type UserCapabilityState struct {
	UserID       uuid.UUID       `json:"user_id" db:"user_id"`
	CapabilityKey string         `json:"capability_key" db:"capability_key"`
	Enrolled     bool            `json:"enrolled" db:"enrolled"`
	StateData    json.RawMessage `json:"state_data,omitempty" db:"state_data"`
	EnrolledAt   *time.Time      `json:"enrolled_at,omitempty" db:"enrolled_at"`
	LastUsedAt   *time.Time      `json:"last_used_at,omitempty" db:"last_used_at"`
}

// IsEnrolled checks if the user is enrolled in the capability
func (ucs *UserCapabilityState) IsEnrolled() bool {
	return ucs.Enrolled
}

// GetStateData returns the state data as a map
func (ucs *UserCapabilityState) GetStateData() (map[string]interface{}, error) {
	if len(ucs.StateData) == 0 {
		return make(map[string]interface{}), nil
	}
	var result map[string]interface{}
	err := json.Unmarshal(ucs.StateData, &result)
	return result, err
}

// SetStateData sets the state data from a map
func (ucs *UserCapabilityState) SetStateData(data map[string]interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	ucs.StateData = jsonData
	return nil
}

// GetTOTPSecret returns the TOTP secret from state data
func (ucs *UserCapabilityState) GetTOTPSecret() (string, error) {
	stateData, err := ucs.GetStateData()
	if err != nil {
		return "", err
	}
	if secret, ok := stateData["secret"].(string); ok {
		return secret, nil
	}
	return "", nil
}

// SetTOTPSecret sets the TOTP secret in state data
func (ucs *UserCapabilityState) SetTOTPSecret(secret string) error {
	stateData, err := ucs.GetStateData()
	if err != nil {
		return err
	}
	stateData["secret"] = secret
	return ucs.SetStateData(stateData)
}

// GetRecoveryCodes returns recovery codes from state data
func (ucs *UserCapabilityState) GetRecoveryCodes() ([]string, error) {
	stateData, err := ucs.GetStateData()
	if err != nil {
		return nil, err
	}
	if codes, ok := stateData["recovery_codes"].([]interface{}); ok {
		result := make([]string, 0, len(codes))
		for _, code := range codes {
			if str, ok := code.(string); ok {
				result = append(result, str)
			}
		}
		return result, nil
	}
	return nil, nil
}

// SetRecoveryCodes sets recovery codes in state data
func (ucs *UserCapabilityState) SetRecoveryCodes(codes []string) error {
	stateData, err := ucs.GetStateData()
	if err != nil {
		return err
	}
	stateData["recovery_codes"] = codes
	return ucs.SetStateData(stateData)
}

