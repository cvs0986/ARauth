package totp

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"image/png"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// Generator provides TOTP generation functionality
type Generator struct {
	issuer string
}

// NewGenerator creates a new TOTP generator
func NewGenerator(issuer string) *Generator {
	return &Generator{issuer: issuer}
}

// GenerateSecret generates a new TOTP secret
func (g *Generator) GenerateSecret(accountName string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      g.issuer,
		AccountName: accountName,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	return key.Secret(), nil
}

// GenerateQRCode generates a QR code for TOTP setup
func (g *Generator) GenerateQRCode(accountName string, secret string) ([]byte, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      g.issuer,
		AccountName: accountName,
		Secret:      []byte(secret),
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	// Generate QR code
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code image: %w", err)
	}

	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode QR code: %w", err)
	}

	return buf.Bytes(), nil
}

// Validate validates a TOTP code
func (g *Generator) Validate(secret string, code string) bool {
	return totp.Validate(code, secret)
}

// GenerateRecoveryCodes generates recovery codes for MFA
func (g *Generator) GenerateRecoveryCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		// Generate 16-character recovery code
		code := generateRandomCode(16)
		codes[i] = code
	}
	return codes, nil
}

// generateRandomCode generates a random alphanumeric code
func generateRandomCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	
	// Generate random bytes
	randomBytes := make([]byte, 4)
	for i := range b {
		rand.Read(randomBytes)
		idx := binary.BigEndian.Uint32(randomBytes) % uint32(len(charset))
		b[i] = charset[idx]
	}
	
	return string(b)
}

