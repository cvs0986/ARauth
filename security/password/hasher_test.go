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
	err = hasher.Verify(password, hash)
	assert.NoError(t, err)

	// Verify incorrect password
	err = hasher.Verify("WrongPassword", hash)
	assert.Error(t, err)
}

func TestHasher_Verify_InvalidHash(t *testing.T) {
	hasher := NewHasher()
	password := "TestPassword123!"

	err := hasher.Verify(password, "invalid-hash")
	assert.Error(t, err)
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
	err := hasher.Verify(password, hash1)
	assert.NoError(t, err)

	err = hasher.Verify(password, hash2)
	assert.NoError(t, err)
}

