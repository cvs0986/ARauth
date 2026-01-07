package encryption

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptor_EncryptDecrypt(t *testing.T) {
	key := []byte("01234567890123456789012345678901") // 32 bytes for AES-256
	encryptor, err := NewEncryptor(key)
	require.NoError(t, err)

	plaintext := "sensitive data"

	encrypted, err := encryptor.Encrypt(plaintext)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := encryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptor_Encrypt_InvalidKey(t *testing.T) {
	key := []byte("short") // Too short
	_, err := NewEncryptor(key)
	assert.Error(t, err)
}

func TestEncryptor_Decrypt_InvalidData(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	encryptor, err := NewEncryptor(key)
	require.NoError(t, err)

	_, err = encryptor.Decrypt("invalid-encrypted-data")
	assert.Error(t, err)
}

func TestEncryptor_Consistency(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	encryptor, err := NewEncryptor(key)
	require.NoError(t, err)

	plaintext := "test data"

	encrypted1, err1 := encryptor.Encrypt(plaintext)
	require.NoError(t, err1)

	encrypted2, err2 := encryptor.Encrypt(plaintext)
	require.NoError(t, err2)

	// Encrypted values should be different (due to nonce)
	assert.NotEqual(t, encrypted1, encrypted2)

	// But both should decrypt to the same value
	decrypted1, err := encryptor.Decrypt(encrypted1)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted1)

	decrypted2, err := encryptor.Decrypt(encrypted2)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted2)
}

