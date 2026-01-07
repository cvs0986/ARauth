package totp

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

// BenchmarkGenerateSecret benchmarks TOTP secret generation
func BenchmarkGenerateSecret(b *testing.B) {
	generator := NewGenerator("Test Issuer")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.GenerateSecret("test@example.com")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValidate benchmarks TOTP code validation
func BenchmarkValidate(b *testing.B) {
	generator := NewGenerator("Test Issuer")
	secret, err := generator.GenerateSecret("test@example.com")
	if err != nil {
		b.Fatal(err)
	}

	// Generate a valid code for testing using the totp library directly
	import "github.com/pquerna/otp/totp"
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.Validate(secret, code)
	}
}

// BenchmarkGenerateQRCode benchmarks QR code generation
func BenchmarkGenerateQRCode(b *testing.B) {
	generator := NewGenerator("Test Issuer")
	secret, err := generator.GenerateSecret("test@example.com")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.GenerateQRCode("test@example.com", secret)
		if err != nil {
			b.Fatal(err)
		}
	}
}

