// Package config handles loading and validating application configuration
// from environment variables and optional YAML files using koanf.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config is the root configuration struct for SaaSKit.
type Config struct {
	Env     string `koanf:"env"`
	Port    int    `koanf:"port"`
	BaseURL string `koanf:"base_url"`

	Database DatabaseConfig `koanf:"database"`
	JWT      JWTConfig      `koanf:"jwt"`
	Argon2   Argon2Config   `koanf:"argon2"`
	OAuth    OAuthConfig    `koanf:"oauth"`
	Log      LogConfig      `koanf:"log"`

	ServerSecret        string `koanf:"server_secret"`
	EncryptionMasterKey string `koanf:"encryption_master_key"`

	RateLimit RateLimitConfig `koanf:"rate_limit"`
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	User     string `koanf:"user"`
	Password string `koanf:"password"`
	Name     string `koanf:"name"`
	SSLMode  string `koanf:"sslmode"`
	MaxConns int32  `koanf:"max_conns"`
	MinConns int32  `koanf:"min_conns"`
}

// DSN builds a PostgreSQL connection string from individual fields.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

// JWTConfig holds JWT signing configuration.
type JWTConfig struct {
	Algorithm  string        `koanf:"algorithm"`
	KeyPath    string        `koanf:"key_path"`
	Issuer     string        `koanf:"issuer"`
	AccessTTL  time.Duration `koanf:"access_ttl"`
	RefreshTTL time.Duration `koanf:"refresh_ttl"`
}

// Argon2Config holds Argon2id password hashing parameters.
type Argon2Config struct {
	Memory      uint32 `koanf:"memory"`
	Iterations  uint32 `koanf:"iterations"`
	Parallelism uint8  `koanf:"parallelism"`
	SaltLength  uint32 `koanf:"salt_length"`
	KeyLength   uint32 `koanf:"key_length"`
}

// OAuthConfig holds OAuth provider credentials.
type OAuthConfig struct {
	Google OAuthProviderConfig `koanf:"google"`
	GitHub OAuthProviderConfig `koanf:"github"`
}

// OAuthProviderConfig holds credentials for a single OAuth provider.
type OAuthProviderConfig struct {
	ClientID     string `koanf:"client_id"`
	ClientSecret string `koanf:"client_secret"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level  string `koanf:"level"`
	Format string `koanf:"format"`
}

// RateLimitConfig holds rate limiting settings.
type RateLimitConfig struct {
	AuthRPM int `koanf:"auth_rpm"`
	APIRPM  int `koanf:"api_rpm"`
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Env == "development" || c.Env == "dev"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.Env == "production" || c.Env == "prod"
}

// Load reads configuration from environment variables (prefix SAASKIT_)
// and an optional YAML file. Env vars take precedence over file values.
func Load(configPath string) (*Config, error) {
	k := koanf.New(".")

	// Defaults
	defaults := map[string]interface{}{
		"env":                 "development",
		"port":                8080,
		"base_url":            "http://localhost:8080",
		"database.host":       "localhost",
		"database.port":       5432,
		"database.user":       "saaskit",
		"database.password":   "saaskit",
		"database.name":       "saaskit",
		"database.sslmode":    "disable",
		"database.max_conns":  25,
		"database.min_conns":  5,
		"jwt.algorithm":       "RS256",
		"jwt.key_path":        "./keys",
		"jwt.issuer":          "http://localhost:8080",
		"jwt.access_ttl":      15 * time.Minute,
		"jwt.refresh_ttl":     7 * 24 * time.Hour,
		"argon2.memory":       65536,
		"argon2.iterations":   3,
		"argon2.parallelism":  4,
		"argon2.salt_length":  16,
		"argon2.key_length":   32,
		"log.level":           "info",
		"log.format":          "json",
		"rate_limit.auth_rpm": 10,
		"rate_limit.api_rpm":  60,
	}
	for key, val := range defaults {
		_ = k.Set(key, val)
	}

	// Optional YAML config file
	if configPath != "" {
		if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
			// Non-fatal if file doesn't exist
			_ = err
		}
	}

	// Environment variables (SAASKIT_ prefix, e.g. SAASKIT_DATABASE_URL → database.url)
	err := k.Load(env.Provider("SAASKIT_", ".", func(s string) string {
		return strings.ReplaceAll(
			strings.ToLower(strings.TrimPrefix(s, "SAASKIT_")),
			"_", ".",
		)
	}), nil)
	if err != nil {
		return nil, fmt.Errorf("loading env vars: %w", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("SAASKIT_DATABASE_HOST is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("SAASKIT_DATABASE_USER is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("SAASKIT_DATABASE_NAME is required")
	}
	if c.IsProduction() {
		if c.ServerSecret == "" || c.ServerSecret == "change-me-in-production-use-64-random-bytes" {
			return fmt.Errorf("SAASKIT_SERVER_SECRET must be set to a secure value in production")
		}
		if c.EncryptionMasterKey == "" || c.EncryptionMasterKey == "change-me-32-byte-hex-encoded-key" {
			return fmt.Errorf("SAASKIT_ENCRYPTION_MASTER_KEY must be set in production")
		}
	}
	validAlgorithms := map[string]bool{"RS256": true, "ES256": true, "EdDSA": true}
	if !validAlgorithms[c.JWT.Algorithm] {
		return fmt.Errorf("SAASKIT_JWT_ALGORITHM must be RS256, ES256, or EdDSA (got %q)", c.JWT.Algorithm)
	}
	return nil
}
