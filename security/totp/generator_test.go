package totp

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_GenerateSecret(t *testing.T) {
	generator := NewGenerator("TestApp")

	secret, err := generator.GenerateSecret("test@example.com")
	require.NoError(t, err)
	assert.NotEmpty(t, secret)
	assert.Len(t, secret, 32) // Base32 encoded secret length
}

func TestGenerator_GenerateQRCode(t *testing.T) {
	generator := NewGenerator("TestApp")

	secret, err := generator.GenerateSecret("test@example.com")
	require.NoError(t, err)

	qrCode, err := generator.GenerateQRCode(secret, "test@example.com")
	require.NoError(t, err)
	assert.NotEmpty(t, qrCode)
}

func TestGenerator_ValidateCode(t *testing.T) {
	generator := NewGenerator("TestApp")

	secret, err := generator.GenerateSecret("test@example.com")
	require.NoError(t, err)

	// Generate a valid code using the same secret
	code, err := totp.GenerateCode(secret, time.Now())
	require.NoError(t, err)

	// Validate the code
	valid := generator.Validate(secret, code)
	assert.True(t, valid)

	// Validate invalid code
	valid = generator.Validate(secret, "000000")
	assert.False(t, valid)
}

func TestGenerator_ValidateCode_InvalidSecret(t *testing.T) {
	generator := NewGenerator("TestApp")

	valid := generator.Validate("invalid-secret", "123456")
	assert.False(t, valid)
}

