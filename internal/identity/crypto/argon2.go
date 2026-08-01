// Package crypto provides password hashing (Argon2id) and JWT key management
// for the identity module.
package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Params holds the Argon2id hashing parameters.
type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultArgon2Params returns OWASP-recommended Argon2id parameters.
func DefaultArgon2Params() Argon2Params {
	return Argon2Params{
		Memory:      64 * 1024, // 64 MiB
		Iterations:  3,
		Parallelism: 4,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// Hasher provides Argon2id password hashing and verification.
type Hasher struct {
	params Argon2Params
}

// NewHasher creates a new password hasher with the given parameters.
func NewHasher(params Argon2Params) *Hasher {
	return &Hasher{params: params}
}

// Hash generates an Argon2id hash of the password in PHC string format:
// $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
func (h *Hasher) Hash(password string) (string, error) {
	salt := make([]byte, h.params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.params.Iterations,
		h.params.Memory,
		h.params.Parallelism,
		h.params.KeyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.params.Memory, h.params.Iterations,
		h.params.Parallelism, b64Salt, b64Hash,
	), nil
}

// Verify checks whether the provided password matches the PHC-encoded hash.
// Uses constant-time comparison to prevent timing attacks.
func (h *Hasher) Verify(password, encodedHash string) (bool, error) {
	params, salt, hash, err := decodePHC(encodedHash)
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	return subtle.ConstantTimeCompare(hash, otherHash) == 1, nil
}

// decodePHC parses a PHC string format hash.
func decodePHC(encodedHash string) (params Argon2Params, salt, hash []byte, err error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return params, nil, nil, fmt.Errorf("invalid hash format: expected 6 parts, got %d", len(parts))
	}

	if parts[1] != "argon2id" {
		return params, nil, nil, fmt.Errorf("unsupported algorithm: %s", parts[1])
	}

	var version int
	_, err = fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return params, nil, nil, fmt.Errorf("parsing version: %w", err)
	}

	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d",
		&params.Memory, &params.Iterations, &params.Parallelism)
	if err != nil {
		return params, nil, nil, fmt.Errorf("parsing params: %w", err)
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return params, nil, nil, fmt.Errorf("decoding salt: %w", err)
	}

	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return params, nil, nil, fmt.Errorf("decoding hash: %w", err)
	}
	params.KeyLength = uint32(len(hash))
	params.SaltLength = uint32(len(salt))

	return params, salt, hash, nil
}
