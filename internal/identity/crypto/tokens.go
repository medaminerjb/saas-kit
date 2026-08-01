package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// TokenHasher provides HMAC-based token hashing.
// Tokens (refresh tokens, identity tokens) are never stored in plaintext.
// Instead, HMAC(server_secret, token) is stored.
// This means a database leak alone cannot be used to validate tokens
// without also having the server secret.
type TokenHasher struct {
	serverSecret []byte
}

// NewTokenHasher creates a new token hasher with the given server secret.
func NewTokenHasher(serverSecret string) *TokenHasher {
	return &TokenHasher{serverSecret: []byte(serverSecret)}
}

// GenerateToken generates a cryptographically random token and returns
// both the raw token (to send to the client) and its HMAC hash (to store).
func (h *TokenHasher) GenerateToken(size int) (raw string, hash string, err error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generating random bytes: %w", err)
	}

	raw = base64.URLEncoding.EncodeToString(b)
	hash = h.Hash(raw)
	return raw, hash, nil
}

// Hash computes HMAC-SHA256(server_secret, token) and returns the hex-encoded result.
func (h *TokenHasher) Hash(token string) string {
	mac := hmac.New(sha256.New, h.serverSecret)
	mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify checks if a token matches the stored hash using constant-time comparison.
func (h *TokenHasher) Verify(token, storedHash string) bool {
	computed := h.Hash(token)
	return hmac.Equal([]byte(computed), []byte(storedHash))
}
