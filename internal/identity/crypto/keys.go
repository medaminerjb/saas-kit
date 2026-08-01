package crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

// KeyPair holds a private/public key pair for JWT signing.
type KeyPair struct {
	PrivateKey crypto.Signer
	PublicKey  crypto.PublicKey
	Algorithm string // RS256, ES256, EdDSA
	KeyID     string // kid for JWKS
}

// LoadOrGenerateKeyPair loads signing keys from disk.
// In development mode, generates keys if they don't exist.
// In production mode, returns an error if keys are missing.
func LoadOrGenerateKeyPair(keyPath, algorithm string, isDev bool) (*KeyPair, error) {
	privPath := filepath.Join(keyPath, "active.pem")
	pubPath := filepath.Join(keyPath, "active.pub")

	// Try to load existing keys
	kp, err := loadKeyPair(privPath, pubPath, algorithm)
	if err == nil {
		return kp, nil
	}

	if !isDev {
		return nil, fmt.Errorf("signing keys not found at %s — in production, keys must be provisioned manually (see keys/README.md): %w", keyPath, err)
	}

	// Development: auto-generate keys
	fmt.Printf("⚠️  Generating %s signing keys for development (path: %s)\n", algorithm, keyPath)
	if err := os.MkdirAll(keyPath, 0700); err != nil {
		return nil, fmt.Errorf("creating key directory: %w", err)
	}

	return generateKeyPair(privPath, pubPath, algorithm)
}

func loadKeyPair(privPath, pubPath, algorithm string) (*KeyPair, error) {
	privPEM, err := os.ReadFile(privPath)
	if err != nil {
		return nil, fmt.Errorf("reading private key: %w", err)
	}

	block, _ := pem.Decode(privPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from %s", privPath)
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Fallback: try legacy PKCS1 for RSA
		rsaKey, rsaErr := x509.ParsePKCS1PrivateKey(block.Bytes)
		if rsaErr != nil {
			// Fallback: try EC
			ecKey, ecErr := x509.ParseECPrivateKey(block.Bytes)
			if ecErr != nil {
				return nil, fmt.Errorf("parsing private key (tried PKCS8, PKCS1, EC): %w", err)
			}
			privKey = ecKey
		} else {
			privKey = rsaKey
		}
	}

	signer, ok := privKey.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("private key does not implement crypto.Signer")
	}

	return &KeyPair{
		PrivateKey: signer,
		PublicKey:  signer.Public(),
		Algorithm:  algorithm,
		KeyID:      "active",
	}, nil
}

func generateKeyPair(privPath, pubPath, algorithm string) (*KeyPair, error) {
	var privKey crypto.Signer
	var err error

	switch algorithm {
	case "RS256":
		privKey, err = rsa.GenerateKey(rand.Reader, 4096)
	case "ES256":
		privKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "EdDSA":
		_, privKey, err = ed25519.GenerateKey(rand.Reader)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
	if err != nil {
		return nil, fmt.Errorf("generating %s key: %w", algorithm, err)
	}

	// Marshal private key to PKCS8
	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, fmt.Errorf("marshalling private key: %w", err)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})
	if err := os.WriteFile(privPath, privPEM, 0600); err != nil {
		return nil, fmt.Errorf("writing private key: %w", err)
	}

	// Marshal public key
	pubBytes, err := x509.MarshalPKIXPublicKey(privKey.Public())
	if err != nil {
		return nil, fmt.Errorf("marshalling public key: %w", err)
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})
	if err := os.WriteFile(pubPath, pubPEM, 0644); err != nil {
		return nil, fmt.Errorf("writing public key: %w", err)
	}

	return &KeyPair{
		PrivateKey: privKey,
		PublicKey:  privKey.Public(),
		Algorithm:  algorithm,
		KeyID:      "active",
	}, nil
}
