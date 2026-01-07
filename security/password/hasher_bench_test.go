package password

import (
	"testing"
)

// BenchmarkHash benchmarks password hashing performance
func BenchmarkHash(b *testing.B) {
	hasher := NewHasher()
	password := "SecurePassword123!@#$"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hasher.Hash(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkVerify benchmarks password verification performance
func BenchmarkVerify(b *testing.B) {
	hasher := NewHasher()
	password := "SecurePassword123!@#$"
	hash, err := hasher.Hash(password)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hasher.Verify(password, hash)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkHashLongPassword benchmarks hashing with longer passwords
func BenchmarkHashLongPassword(b *testing.B) {
	hasher := NewHasher()
	password := "ThisIsAVeryLongPasswordThatExceedsTypicalLengthRequirements123!@#$%^&*()"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hasher.Hash(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}
