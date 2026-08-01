package crypto

import (
	"strings"
	"testing"
)

func TestHasher_HashAndVerify(t *testing.T) {
	hasher := NewHasher(Argon2Params{
		Memory:      64 * 1024,
		Iterations:  1, // Fast for tests
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	})

	password := "SuperSecret123!"

	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Hash() error: %v", err)
	}

	// Verify format
	if !strings.HasPrefix(hash, "$argon2id$v=") {
		t.Errorf("hash should start with $argon2id$v=, got: %s", hash)
	}

	// Correct password should verify
	ok, err := hasher.Verify(password, hash)
	if err != nil {
		t.Fatalf("Verify() error: %v", err)
	}
	if !ok {
		t.Error("Verify() returned false for correct password")
	}

	// Wrong password should not verify
	ok, err = hasher.Verify("wrong-password", hash)
	if err != nil {
		t.Fatalf("Verify() error: %v", err)
	}
	if ok {
		t.Error("Verify() returned true for wrong password")
	}
}

func TestHasher_DifferentSalts(t *testing.T) {
	hasher := NewHasher(Argon2Params{
		Memory:      64 * 1024,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	})

	hash1, _ := hasher.Hash("password")
	hash2, _ := hasher.Hash("password")

	if hash1 == hash2 {
		t.Error("two hashes of the same password should differ (different salts)")
	}
}

func TestDecodePHC_InvalidFormat(t *testing.T) {
	tests := []struct {
		name string
		hash string
	}{
		{"empty", ""},
		{"no dollar signs", "argon2id"},
		{"wrong algorithm", "$bcrypt$v=19$m=65536,t=3,p=4$salt$hash"},
		{"too few parts", "$argon2id$v=19$m=65536"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewHasher(DefaultArgon2Params())
			_, err := hasher.Verify("password", tt.hash)
			if err == nil {
				t.Error("expected error for invalid hash format")
			}
		})
	}
}
