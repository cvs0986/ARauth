package totp

import (
	"testing"

	"github.com/pquerna/otp"
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
	// Use time.Now() directly since ValidateCode uses it internally
	code, err := otp.GenerateCode(secret, nil)
	require.NoError(t, err)

	// Validate the code
	valid := generator.ValidateCode(secret, code)
	assert.True(t, valid)

	// Validate invalid code
	valid = generator.ValidateCode(secret, "000000")
	assert.False(t, valid)
}

func TestGenerator_ValidateCode_InvalidSecret(t *testing.T) {
	generator := NewGenerator("TestApp")

	valid := generator.ValidateCode("invalid-secret", "123456")
	assert.False(t, valid)
}

