package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasher_Hash(t *testing.T) {
	hasher := NewHasher()
	password := "TestPassword123!"

	hash, err := hasher.Hash(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestHasher_Verify(t *testing.T) {
	hasher := NewHasher()
	password := "TestPassword123!"

	hash, err := hasher.Hash(password)
	require.NoError(t, err)

	// Verify correct password
	valid, err := hasher.Verify(password, hash)
	require.NoError(t, err)
	assert.True(t, valid)

	// Verify incorrect password
	valid, err = hasher.Verify("WrongPassword", hash)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestHasher_Verify_InvalidHash(t *testing.T) {
	hasher := NewHasher()
	password := "TestPassword123!"

	valid, err := hasher.Verify(password, "invalid-hash")
	assert.Error(t, err)
	assert.False(t, valid)
}

func TestHasher_Consistency(t *testing.T) {
	hasher := NewHasher()
	password := "TestPassword123!"

	hash1, err1 := hasher.Hash(password)
	require.NoError(t, err1)

	hash2, err2 := hasher.Hash(password)
	require.NoError(t, err2)

	// Hashes should be different (due to salt)
	assert.NotEqual(t, hash1, hash2)

	// But both should verify correctly
	valid, err := hasher.Verify(password, hash1)
	require.NoError(t, err)
	assert.True(t, valid)

	valid, err = hasher.Verify(password, hash2)
	require.NoError(t, err)
	assert.True(t, valid)
}

