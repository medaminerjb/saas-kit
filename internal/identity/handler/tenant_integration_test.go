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
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	idcrypto "github.com/medaminerjb/saas-kit/internal/identity/crypto"
	"github.com/medaminerjb/saas-kit/internal/identity/repository"
	"github.com/medaminerjb/saas-kit/internal/identity/service"
	"github.com/medaminerjb/saas-kit/internal/platform/events"
	tenantrepo "github.com/medaminerjb/saas-kit/internal/tenant/repository"
	tenantservice "github.com/medaminerjb/saas-kit/internal/tenant/service"
)

func TestIntegration_MultiTenancy_E2E(t *testing.T) {
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

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	// Clean database slate and run all migrations
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
	tenantRepo := tenantrepo.NewTenantRepository(pool)

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
	tenantService := tenantservice.NewTenantService(tenantRepo, publisher, slog.Default())

	// Initialize server router
	appHandler := NewRouter(RouterConfig{
		Identity:      identityManager,
		Pool:          pool,
		Logger:        slog.Default(),
		TenantService: tenantService,
	})

	server := httptest.NewServer(appHandler)
	defer server.Close()

	client := &http.Client{}

	// Helper function for registration and login
	registerAndLogin := func(email, password, name string) string {
		// Register
		regBody, _ := json.Marshal(map[string]string{
			"email":    email,
			"password": password,
			"name":     name,
		})
		resp, err := client.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(regBody))
		if err != nil {
			t.Fatalf("failed to register: %v", err)
		}
		_ = resp.Body.Close()

		// Login
		loginBody, _ := json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})
		resp, err = client.Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(loginBody))
		if err != nil {
			t.Fatalf("failed to login: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		var result map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&result)
		tokens := result["tokens"].(map[string]any)
		return tokens["access_token"].(string)
	}

	// 1. Authenticate two test users
	userAToken := registerAndLogin("userA@saaskit.test", "SecureP@ss123", "User A")
	userBToken := registerAndLogin("userB@saaskit.test", "SecureP@ss123", "User B")

	var tenantID string
	var inviteToken string
	var userBID string

	// Get User B's ID to verify membership later
	t.Run("Get User B Profile", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/v1/users/me", nil)
		req.Header.Set("Authorization", "Bearer "+userBToken)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		var profile map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&profile)
		userBID = profile["id"].(string)
	})

	// 2. Create Tenant (User A)
	t.Run("Create Tenant", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"name": "Acme Corporation",
			"slug": "acme",
		})
		req, _ := http.NewRequest("POST", server.URL+"/api/v1/tenants", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+userAToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusCreated {
			b, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 201, got %d. Body: %s", resp.StatusCode, b)
		}

		var tenant map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&tenant)
		tenantID = tenant["id"].(string)
		if tenant["name"] != "Acme Corporation" || tenant["slug"] != "acme" {
			t.Errorf("unexpected tenant details: %+v", tenant)
		}
	})

	// 3. List Tenants (User A)
	t.Run("List Tenants for User A", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/v1/tenants", nil)
		req.Header.Set("Authorization", "Bearer "+userAToken)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		var res map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&res)
		tenants := res["tenants"].([]any)
		if len(tenants) != 1 {
			t.Fatalf("expected 1 tenant, got %d", len(tenants))
		}
		item := tenants[0].(map[string]any)
		if item["role"] != "owner" {
			t.Errorf("expected role 'owner', got %q", item["role"])
		}
	})

	// 4. Update Tenant Settings (User A)
	t.Run("Update Tenant Settings by Owner", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"name": "Acme Inc",
		})
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("%s/api/v1/tenants/%s", server.URL, tenantID), bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+userAToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var tenant map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&tenant)
		if tenant["name"] != "Acme Inc" {
			t.Errorf("expected updated name 'Acme Inc', got %q", tenant["name"])
		}
	})

	// 5. Update Tenant Settings (User B - Unauthorized)
	t.Run("Update Tenant Settings by Non-Member", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"name": "Hacked Inc",
		})
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("%s/api/v1/tenants/%s", server.URL, tenantID), bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+userBToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden, got %d", resp.StatusCode)
		}
	})

	// 6. Invite User B (User A)
	t.Run("Invite Member to Tenant", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"email": "userB@saaskit.test",
			"role":  "member",
		})
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/tenants/%s/members", server.URL, tenantID), bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+userAToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusCreated {
			b, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 201 Created, got %d. Body: %s", resp.StatusCode, b)
		}

		var invite map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&invite)
		inviteToken = invite["token"].(string)
		if inviteToken == "" {
			t.Fatal("expected invite token to be returned in body")
		}
	})

	// 7. Accept Invitation (User B)
	t.Run("Accept Invitation", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"token": inviteToken,
		})
		req, _ := http.NewRequest("POST", server.URL+"/api/v1/tenants/invitations/accept", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+userBToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200 OK, got %d. Body: %s", resp.StatusCode, b)
		}
	})

	// 8. List Members (User B)
	t.Run("List Members as Member", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/tenants/%s/members", server.URL, tenantID), nil)
		req.Header.Set("Authorization", "Bearer "+userBToken)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}

		var res map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&res)
		members := res["members"].([]any)
		if len(members) != 2 {
			t.Errorf("expected 2 members, got %d", len(members))
		}
	})

	// 9. Remove Member (User A removes User B)
	t.Run("Remove Member", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/tenants/%s/members/%s", server.URL, tenantID, userBID), nil)
		req.Header.Set("Authorization", "Bearer "+userAToken)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200 OK, got %d. Body: %s", resp.StatusCode, b)
		}
	})

	// 10. Access check for User B after removal
	t.Run("Access Tenant After Removal", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/tenants/%s", server.URL, tenantID), nil)
		req.Header.Set("Authorization", "Bearer "+userBToken)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden, got %d", resp.StatusCode)
		}
	})
}

