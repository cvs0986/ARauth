package encryption

import (
	"testing"
)

// BenchmarkEncrypt benchmarks encryption performance
func BenchmarkEncrypt(b *testing.B) {
	key := make([]byte, 32)
	copy(key, "test-encryption-key-32-bytes!!")
	encryptor, err := NewEncryptor(key)
	if err != nil {
		b.Fatal(err)
	}

	plaintext := "This is a test secret that needs to be encrypted"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encryptor.Encrypt(plaintext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDecrypt benchmarks decryption performance
func BenchmarkDecrypt(b *testing.B) {
	key := make([]byte, 32)
	copy(key, "test-encryption-key-32-bytes!!")
	encryptor, err := NewEncryptor(key)
	if err != nil {
		b.Fatal(err)
	}

	plaintext := "This is a test secret that needs to be encrypted"
	ciphertext, err := encryptor.Encrypt(plaintext)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encryptor.Decrypt(ciphertext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEncryptDecrypt benchmarks full encrypt-decrypt cycle
func BenchmarkEncryptDecrypt(b *testing.B) {
	key := make([]byte, 32)
	copy(key, "test-encryption-key-32-bytes!!")
	encryptor, err := NewEncryptor(key)
	if err != nil {
		b.Fatal(err)
	}

	plaintext := "This is a test secret that needs to be encrypted"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ciphertext, err := encryptor.Encrypt(plaintext)
		if err != nil {
			b.Fatal(err)
		}
		_, err = encryptor.Decrypt(ciphertext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

