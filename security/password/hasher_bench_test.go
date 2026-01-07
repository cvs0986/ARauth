package password

import (
	"testing"
)

func BenchmarkHasher_Hash(b *testing.B) {
	hasher := NewHasher()
	password := "TestPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hasher.Hash(password)
	}
}

func BenchmarkHasher_Verify(b *testing.B) {
	hasher := NewHasher()
	password := "TestPassword123!"
	hash, _ := hasher.Hash(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hasher.Verify(password, hash)
	}
}

