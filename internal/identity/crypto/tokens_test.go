package crypto

import (
	"testing"
)

func TestTokenHasher_GenerateAndVerify(t *testing.T) {
	hasher := NewTokenHasher("test-server-secret-64-bytes-long-for-hmac-security-000000000")

	raw, hash, err := hasher.GenerateToken(32)
	if err != nil {
		t.Fatalf("GenerateToken() error: %v", err)
	}

	if raw == "" {
		t.Error("raw token should not be empty")
	}
	if hash == "" {
		t.Error("hash should not be empty")
	}
	if raw == hash {
		t.Error("raw token and hash should differ")
	}

	// Verify the generated token matches its hash
	if !hasher.Verify(raw, hash) {
		t.Error("Verify() should return true for matching token and hash")
	}

	// Wrong token should not verify
	if hasher.Verify("wrong-token", hash) {
		t.Error("Verify() should return false for wrong token")
	}
}

func TestTokenHasher_DifferentSecrets(t *testing.T) {
	hasher1 := NewTokenHasher("secret-one-000000000000000000000000000000000000000000000")
	hasher2 := NewTokenHasher("secret-two-000000000000000000000000000000000000000000000")

	token := "some-refresh-token"

	hash1 := hasher1.Hash(token)
	hash2 := hasher2.Hash(token)

	if hash1 == hash2 {
		t.Error("same token with different secrets should produce different hashes")
	}

	// Can't verify with wrong secret
	if hasher2.Verify(token, hash1) {
		t.Error("should not verify with different server secret")
	}
}

func TestTokenHasher_Deterministic(t *testing.T) {
	hasher := NewTokenHasher("consistent-secret-for-testing-purposes-00000000000000000")

	token := "my-refresh-token"
	hash1 := hasher.Hash(token)
	hash2 := hasher.Hash(token)

	if hash1 != hash2 {
		t.Error("same token with same secret should produce identical hashes")
	}
}
