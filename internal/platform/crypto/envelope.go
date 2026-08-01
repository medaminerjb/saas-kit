// Package crypto provides envelope encryption (AES-256-GCM) for secrets stored in the database.
//
// Every stored secret (OAuth client secrets, SMTP passwords, webhook secrets, etc.)
// is encrypted with a data encryption key (DEK) derived from the master key using HKDF.
// This ensures that if the database is compromised, secrets remain protected.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

// Envelope provides AES-256-GCM envelope encryption using a master key.
type Envelope struct {
	masterKey []byte
}

// NewEnvelope creates a new Envelope encryptor from a hex-encoded master key.
// The master key must be exactly 32 bytes (64 hex characters).
func NewEnvelope(masterKeyHex string) (*Envelope, error) {
	key, err := hex.DecodeString(masterKeyHex)
	if err != nil {
		return nil, fmt.Errorf("decoding master key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("master key must be 32 bytes (64 hex chars), got %d bytes", len(key))
	}
	return &Envelope{masterKey: key}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM with a derived key.
// The context string provides domain separation (e.g., "oauth_client_secret").
// Returns hex-encoded ciphertext (nonce || ciphertext).
func (e *Envelope) Encrypt(plaintext []byte, context string) (string, error) {
	dek, err := e.deriveKey(context)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(dek)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generating nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, []byte(context))
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt decrypts hex-encoded ciphertext produced by Encrypt.
func (e *Envelope) Decrypt(ciphertextHex string, context string) ([]byte, error) {
	data, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return nil, fmt.Errorf("decoding ciphertext: %w", err)
	}

	dek, err := e.deriveKey(context)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, []byte(context))
	if err != nil {
		return nil, fmt.Errorf("decrypting: %w", err)
	}

	return plaintext, nil
}

// deriveKey derives a 32-byte data encryption key from the master key using HKDF-SHA256.
func (e *Envelope) deriveKey(context string) ([]byte, error) {
	r := hkdf.New(sha256.New, e.masterKey, nil, []byte(context))
	dek := make([]byte, 32)
	if _, err := io.ReadFull(r, dek); err != nil {
		return nil, fmt.Errorf("deriving key: %w", err)
	}
	return dek, nil
}
