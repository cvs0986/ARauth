package totp

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"image/png"
	"strings"
	"time"

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
// The secret should be a base32-encoded string (as returned by GenerateSecret)
func (g *Generator) GenerateQRCode(accountName string, secret string) ([]byte, error) {
	// Create a TOTP key from the existing secret using the otpauth URL format
	// This ensures the QR code contains all the necessary information
	key, err := otp.NewKeyFromURL(fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&period=30&digits=6&algorithm=SHA1",
		g.issuer, accountName, secret, g.issuer))
	if err != nil {
		return nil, fmt.Errorf("failed to create TOTP key from secret: %w", err)
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
// Uses ValidateCustom with a time window of ±1 period (30 seconds) to account for clock skew
// The secret should be a base32-encoded string
func (g *Generator) Validate(secret string, code string) bool {
	// Trim whitespace from code
	code = strings.TrimSpace(code)
	
	// Use ValidateCustom to explicitly set time window tolerance
	// This allows codes from the previous and next time windows to be valid
	// The default Validate uses ±1 period, but we make it explicit here
	// Increase skew to 2 to be more tolerant of clock differences
	valid, err := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      2, // Allow ±2 periods (60 seconds) clock skew for better tolerance
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		// Log error for debugging (in production, you might want to use a logger)
		fmt.Printf("TOTP validation error: %v (secret length: %d, code: %s)\n", err, len(secret), code)
		return false
	}
	return valid
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
		if _, err := rand.Read(randomBytes); err != nil {
			// Fallback to deterministic if random fails
			randomBytes[0] = byte(i)
			randomBytes[1] = byte(i >> 8)
			randomBytes[2] = byte(i >> 16)
			randomBytes[3] = byte(i >> 24)
		}
		idx := binary.BigEndian.Uint32(randomBytes) % uint32(len(charset))
		b[i] = charset[idx]
	}
	
	return string(b)
}

