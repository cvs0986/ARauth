package password

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2id parameters
	memory      = 64 * 1024  // 64 MB
	iterations = 3
	parallelism = 4
	saltLength  = 16
	keyLength   = 32
)

// Hasher provides password hashing functionality
type Hasher struct{}

// NewHasher creates a new password hasher
func NewHasher() *Hasher {
	return &Hasher{}
}

// Hash hashes a password using Argon2id
func (h *Hasher) Hash(password string) (string, error) {
	// Generate salt
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash password
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)

	// Encode: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, iterations, parallelism, encodedSalt, encodedHash), nil
}

// Verify verifies a password against a hash
func (h *Hasher) Verify(password string, hash string) (bool, error) {
	// Parse hash
	var version int
	var m, t, p uint32
	var salt, hashBytes []byte

	_, err := fmt.Sscanf(hash, "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		&version, &m, &t, &p, &salt, &hashBytes)
	if err != nil {
		return false, fmt.Errorf("invalid hash format: %w", err)
	}

	// Decode salt and hash
	saltBytes, err := base64.RawStdEncoding.DecodeString(string(salt))
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	hashBytesDecoded, err := base64.RawStdEncoding.DecodeString(string(hashBytes))
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Compute hash with same parameters
	computedHash := argon2.IDKey([]byte(password), saltBytes, t, m, uint8(p), uint32(len(hashBytesDecoded)))

	// Constant-time comparison
	if len(computedHash) != len(hashBytesDecoded) {
		return false, nil
	}

	for i := range computedHash {
		if computedHash[i] != hashBytesDecoded[i] {
			return false, nil
		}
	}

	return true, nil
}

