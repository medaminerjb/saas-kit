package crypto

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrGenerateKeyPair_RS256(t *testing.T) {
	tmpDir := t.TempDir()

	kp, err := LoadOrGenerateKeyPair(tmpDir, "RS256", true)
	if err != nil {
		t.Fatalf("LoadOrGenerateKeyPair() error: %v", err)
	}

	if kp.Algorithm != "RS256" {
		t.Errorf("algorithm = %q, want RS256", kp.Algorithm)
	}
	if kp.PrivateKey == nil {
		t.Error("private key should not be nil")
	}
	if kp.PublicKey == nil {
		t.Error("public key should not be nil")
	}

	// Files should exist
	if _, err := os.Stat(filepath.Join(tmpDir, "active.pem")); os.IsNotExist(err) {
		t.Error("private key file should exist")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "active.pub")); os.IsNotExist(err) {
		t.Error("public key file should exist")
	}

	// Loading again should reuse existing keys
	kp2, err := LoadOrGenerateKeyPair(tmpDir, "RS256", true)
	if err != nil {
		t.Fatalf("second LoadOrGenerateKeyPair() error: %v", err)
	}
	if kp2.PrivateKey == nil {
		t.Error("reloaded private key should not be nil")
	}
}

func TestLoadOrGenerateKeyPair_ES256(t *testing.T) {
	tmpDir := t.TempDir()

	kp, err := LoadOrGenerateKeyPair(tmpDir, "ES256", true)
	if err != nil {
		t.Fatalf("LoadOrGenerateKeyPair(ES256) error: %v", err)
	}
	if kp.Algorithm != "ES256" {
		t.Errorf("algorithm = %q, want ES256", kp.Algorithm)
	}
}

func TestLoadOrGenerateKeyPair_EdDSA(t *testing.T) {
	tmpDir := t.TempDir()

	kp, err := LoadOrGenerateKeyPair(tmpDir, "EdDSA", true)
	if err != nil {
		t.Fatalf("LoadOrGenerateKeyPair(EdDSA) error: %v", err)
	}
	if kp.Algorithm != "EdDSA" {
		t.Errorf("algorithm = %q, want EdDSA", kp.Algorithm)
	}
}

func TestLoadOrGenerateKeyPair_ProductionFailsIfMissing(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := LoadOrGenerateKeyPair(tmpDir, "RS256", false)
	if err == nil {
		t.Error("expected error in production mode when keys are missing")
	}
}
