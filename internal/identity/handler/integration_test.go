package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	idcrypto "github.com/saaskit/saaskit/internal/identity/crypto"
	"github.com/saaskit/saaskit/internal/identity/repository"
	"github.com/saaskit/saaskit/internal/identity/service"
	"github.com/saaskit/saaskit/internal/platform/events"
)

func runMigrations(ctx context.Context, t *testing.T, pool *pgxpool.Pool) {
	migrationsDir := filepath.Join("..", "..", "..", "migrations")
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("failed to read migrations dir: %v", err)
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)

	// Nuke the entire public schema and recreate it — guaranteed clean slate
	_, _ = pool.Exec(ctx, `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`)

	for _, sqlFile := range sqlFiles {
		content, err := os.ReadFile(filepath.Join(migrationsDir, sqlFile))
		if err != nil {
			t.Fatalf("failed to read migration file %s: %v", sqlFile, err)
		}

		lines := strings.Split(string(content), "\n")
		var upSql strings.Builder
		inUp := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "-- +goose Up") {
				inUp = true
				continue
			}
			if strings.HasPrefix(trimmed, "-- +goose Down") {
				inUp = false
				break
			}
			if inUp {
				upSql.WriteString(line)
				upSql.WriteString("\n")
			}
		}

		queries := strings.Split(upSql.String(), ";")
		for _, q := range queries {
			// Strip comment-only lines and goose directives from each chunk
			var cleaned []string
			for _, line := range strings.Split(q, "\n") {
				tl := strings.TrimSpace(line)
				if tl == "" || strings.HasPrefix(tl, "--") {
					continue
				}
				cleaned = append(cleaned, line)
			}
			stmt := strings.TrimSpace(strings.Join(cleaned, "\n"))
			if stmt == "" {
				continue
			}
			_, err = pool.Exec(ctx, stmt)
			if err != nil {
				t.Fatalf("failed to execute migration query from %s:\nQuery: %s\nError: %v", sqlFile, stmt, err)
			}
		}
	}
}

func getTestDSN() string {
	host := os.Getenv("SAASKIT_DATABASE_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("SAASKIT_DATABASE_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("SAASKIT_DATABASE_USER")
	if user == "" {
		user = "saaskit"
	}
	password := os.Getenv("SAASKIT_DATABASE_PASSWORD")
	if password == "" {
		password = "saaskit"
	}
	name := os.Getenv("SAASKIT_DATABASE_NAME")
	if name == "" {
		name = "saaskit_test"
	}
	sslmode := os.Getenv("SAASKIT_DATABASE_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, name, sslmode)
}

func TestIntegration_E2E_Flows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	dsn := getTestDSN()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Verify database connection is alive
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	// Apply database schema migrations
	runMigrations(ctx, t, pool)

	// Set up crypto and keys
	tmpDir := t.TempDir()
	kp, err := idcrypto.LoadOrGenerateKeyPair(tmpDir, "RS256", true)
	if err != nil {
		t.Fatalf("key generation: %v", err)
	}

	tokenService := service.NewTokenService(service.TokenServiceConfig{
		PrivateKey: kp.PrivateKey,
		PublicKey:  kp.PublicKey,
		Algorithm:  "RS256",
		KeyID:      "test-key",
		Issuer:     "http://localhost:8080",
		AccessTTL:  15 * time.Minute,
	}, slog.Default())

	hasher := idcrypto.NewHasher(idcrypto.Argon2Params{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	})

	tokenHasher := idcrypto.NewTokenHasher("ci-test-secret-00000000000000000000000000000000000000000000")

	// Set up repositories
	userRepo := repository.NewUserRepo(pool)
	sessionRepo := repository.NewSessionRepo(pool)
	tokenRepo := repository.NewTokenRepo(pool)

	// Set up services
	logPublisher := events.NewLogPublisher(slog.Default())
	publisher := events.NewMultiPublisher(logPublisher)

	authService := service.NewAuthService(service.AuthServiceConfig{
		Users:        userRepo,
		Sessions:     sessionRepo,
		Tokens:       tokenRepo,
		Hasher:       hasher,
		TokenHasher:  tokenHasher,
		TokenService: tokenService,
		Publisher:    publisher,
		RefreshTTL:   7 * 24 * time.Hour,
	}, slog.Default())

	userService := service.NewUserService(userRepo, publisher, slog.Default())
	identityManager := service.NewIdentityManager(authService, userService, tokenService, slog.Default())

	// Initialize server router
	appHandler := NewRouter(RouterConfig{
		Identity: identityManager,
		Pool:     pool,
		Logger:   slog.Default(),
	})

	server := httptest.NewServer(appHandler)
	defer server.Close()

	client := &http.Client{}

	email := "user@saaskit.test"
	password := "SecureP@ss123"
	name := "Integration Test User"

	// 1. Registration Flow
	t.Run("1. Register User", func(t *testing.T) {
		regBody, _ := json.Marshal(map[string]string{
			"email":    email,
			"password": password,
			"name":     name,
		})

		resp, err := client.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(regBody))
		if err != nil {
			t.Fatalf("failed to send registration request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected registration status 201 Created, got %d. Body: %s", resp.StatusCode, string(body))
		}

		// Verify database has the user
		user, err := userRepo.GetByEmail(ctx, email, nil)
		if err != nil {
			t.Fatalf("failed to retrieve user from DB: %v", err)
		}
		if user.Name != name {
			t.Errorf("expected user name %q, got %q", name, user.Name)
		}
	})

	// 2. Login Flow
	var accessToken, refreshToken string
	t.Run("2. Login User", func(t *testing.T) {
		loginBody, _ := json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})

		resp, err := client.Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(loginBody))
		if err != nil {
			t.Fatalf("failed to send login request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected login status 200 OK, got %d. Body: %s", resp.StatusCode, string(body))
		}

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode login response: %v", err)
		}

		tokens, ok := result["tokens"].(map[string]any)
		if !ok {
			t.Fatalf("tokens missing or invalid in response: %v", result)
		}

		accessToken = tokens["access_token"].(string)
		refreshToken = tokens["refresh_token"].(string)

		if accessToken == "" || refreshToken == "" {
			t.Error("access token or refresh token is empty")
		}
	})

	// 3. Token Refresh Flow
	var newAccessToken, newRefreshToken string
	t.Run("3. Refresh Tokens", func(t *testing.T) {
		refreshBody, _ := json.Marshal(map[string]string{
			"refresh_token": refreshToken,
		})

		resp, err := client.Post(server.URL+"/api/v1/auth/refresh", "application/json", bytes.NewReader(refreshBody))
		if err != nil {
			t.Fatalf("failed to send token refresh request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected refresh status 200 OK, got %d. Body: %s", resp.StatusCode, string(body))
		}

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode refresh response: %v", err)
		}

		newAccessToken = result["access_token"].(string)
		newRefreshToken = result["refresh_token"].(string)

		if newAccessToken == "" || newRefreshToken == "" {
			t.Error("refreshed access token or refresh token is empty")
		}

		if newAccessToken == accessToken || newRefreshToken == refreshToken {
			t.Error("tokens were not rotated")
		}
	})

	// 4. Protected Resource & Logout Flow
	t.Run("4. Logout and Revoke", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/auth/logout", nil)
		if err != nil {
			t.Fatalf("failed to create logout request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+newAccessToken)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to send logout request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected logout status 200 OK, got %d. Body: %s", resp.StatusCode, string(body))
		}

		// 1. Verify in DB that the session is revoked (revoked_at IS NOT NULL)
		var revokedAt *time.Time
		err = pool.QueryRow(ctx, `SELECT revoked_at FROM sessions WHERE revoked_at IS NOT NULL ORDER BY revoked_at DESC LIMIT 1`).Scan(&revokedAt)
		if err != nil {
			t.Fatalf("failed to query revoked session from DB: %v", err)
		}
		if revokedAt == nil {
			t.Error("expected session to be revoked in DB, but revoked_at is NULL")
		}

		// 2. Verify that refreshing with the rotated/revoked refresh token now fails
		refreshBody, _ := json.Marshal(map[string]string{
			"refresh_token": newRefreshToken,
		})
		respRef, err := client.Post(server.URL+"/api/v1/auth/refresh", "application/json", bytes.NewReader(refreshBody))
		if err != nil {
			t.Fatalf("failed to send refresh request after logout: %v", err)
		}
		defer respRef.Body.Close()

		if respRef.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected refresh status 401 Unauthorized for revoked session, got %d", respRef.StatusCode)
		}
	})
}
