package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator_Validate(t *testing.T) {
	validator := NewValidator(12, true, true, true, true)

	tests := []struct {
		name      string
		password  string
		username  string
		wantError bool
	}{
		{
			name:      "valid password",
			password:  "SecurePass123!",
			username:  "testuser",
			wantError: false,
		},
		{
			name:      "too short",
			password:  "Short1!",
			username:  "testuser",
			wantError: true,
		},
		{
			name:      "no uppercase",
			password:  "securepass123!",
			username:  "testuser",
			wantError: true,
		},
		{
			name:      "no lowercase",
			password:  "SECUREPASS123!",
			username:  "testuser",
			wantError: true,
		},
		{
			name:      "no digit",
			password:  "SecurePass!",
			username:  "testuser",
			wantError: true,
		},
		{
			name:      "no special char",
			password:  "SecurePass123",
			username:  "testuser",
			wantError: true,
		},
		{
			name:      "contains username",
			password:  "testuser123!",
			username:  "testuser",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.password, tt.username)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_Validate_CommonPassword(t *testing.T) {
	validator := NewValidator(8, false, false, false, false)

	err := validator.Validate("password", "user")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "common")
}

