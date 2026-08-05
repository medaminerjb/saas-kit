package crypto

import (
	"testing"
)

func FuzzArgon2Hash(f *testing.F) {
	f.Add([]byte("password123"))
	f.Add([]byte("short"))
	f.Add([]byte("a-very-long-password-with-special-chars-!@#$%"))

	hasher := NewHasher(DefaultArgon2Params())

	f.Fuzz(func(t *testing.T, password []byte) {
		if len(password) == 0 || len(password) > 128 {
			return // Skip empty or excessively long passwords
		}

		// Hash the password
		hash, err := hasher.Hash(string(password))
		if err != nil {
			return
		}

		// Verify the password
		_, _ = hasher.Verify(string(password), hash)
	})
}
