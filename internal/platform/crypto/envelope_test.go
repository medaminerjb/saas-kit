package crypto

import (
	"testing"
)

func TestEnvelope_EncryptDecrypt(t *testing.T) {
	// 32 bytes = 64 hex chars
	masterKey := "6368616e676520746869732070617373776f726420746f206120736563726574"

	env, err := NewEnvelope(masterKey)
	if err != nil {
		t.Fatalf("NewEnvelope() error: %v", err)
	}

	plaintext := []byte("my-oauth-client-secret")
	context := "oauth_client_secret"

	ciphertext, err := env.Encrypt(plaintext, context)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	if ciphertext == string(plaintext) {
		t.Error("ciphertext should differ from plaintext")
	}

	decrypted, err := env.Decrypt(ciphertext, context)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
	}
}

func TestEnvelope_WrongContext(t *testing.T) {
	masterKey := "6368616e676520746869732070617373776f726420746f206120736563726574"

	env, err := NewEnvelope(masterKey)
	if err != nil {
		t.Fatalf("NewEnvelope() error: %v", err)
	}

	ciphertext, _ := env.Encrypt([]byte("secret"), "context-a")

	_, err = env.Decrypt(ciphertext, "context-b")
	if err == nil {
		t.Error("Decrypt with wrong context should fail")
	}
}

func TestEnvelope_InvalidMasterKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"too short", "abcdef"},
		{"not hex", "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEnvelope(tt.key)
			if err == nil {
				t.Error("expected error for invalid master key")
			}
		})
	}
}
