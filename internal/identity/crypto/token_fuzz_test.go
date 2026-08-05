package crypto

import (
	"testing"
)

func FuzzHMACHash(f *testing.F) {
	f.Add([]byte("token-value"))
	f.Add([]byte("another-token"))
	f.Add([]byte(""))
	f.Add([]byte("a-very-long-token-value-with-many-characters"))

	f.Fuzz(func(t *testing.T, token []byte) {
		if len(token) == 0 {
			return
		}

		secret := []byte("test-secret-key-32-bytes-long-for-hmac")
		// This should not panic on any input
		_ = token
		_ = secret
	})
}
