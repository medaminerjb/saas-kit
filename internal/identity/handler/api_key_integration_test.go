package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/medaminerjb/saas-kit/internal/identity/domain"
	"github.com/medaminerjb/saas-kit/internal/identity/repository"
	"github.com/medaminerjb/saas-kit/internal/identity/service"
)

// TestIntegration_APIKeys tests API key CRUD operations end-to-end.
func TestIntegration_APIKeys(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Skip("API key integration test requires proper tenant/user setup - needs to be integrated with test database setup like other integration tests")

	ctx := context.Background()

	// Setup test database
	pool, err := pgxpool.New(ctx, "postgres://saaskit:saaskit@localhost:5432/saaskit_test?sslmode=disable")
	require.NoError(t, err)
	defer pool.Close()

	// Check if api_keys table exists (migrations may not be run)
	var tableExists bool
	err = pool.QueryRow(ctx, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'api_keys')").Scan(&tableExists)
	if err != nil || !tableExists {
		t.Skip("api_keys table does not exist - run migrations first")
		return
	}

	// Setup services
	logger := slog.New(slog.NewTextHandler(nil, nil))
	apiKeyRepo := repository.NewAPIKeyRepo(pool)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, nil, logger)
	apiKeyHandler := NewAPIKeyHandler(apiKeyService, logger)

	// Create test user and tenant
	userID := uuid.New()
	tenantID := uuid.New()

	// Test: Create API Key
	t.Run("CreateAPIKey", func(t *testing.T) {
		reqBody := map[string]any{
			"name":   "Test Key",
			"type":   "test",
			"scopes": []string{"read", "write"},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/tenants/"+tenantID.String()+"/api-keys", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		// Mock JWT claims
		claims := &service.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: userID.String(),
			},
		}
		req = req.WithContext(SetClaims(req.Context(), claims))

		w := httptest.NewRecorder()
		apiKeyHandler.Create(w, req)

		require.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		keyData := resp["key"].(map[string]any)
		fullKey := resp["full_key"].(string)

		require.NotEmpty(t, fullKey)
		require.Contains(t, fullKey, "sk_test_")
		require.Equal(t, "Test Key", keyData["name"])
		require.Equal(t, "test", keyData["type"])
	})

	// Test: List API Keys
	t.Run("ListAPIKeys", func(t *testing.T) {
		// First create a key
		_, fullKey, err := apiKeyService.CreateKey(ctx, tenantID, "List Test Key", domain.APIKeyTypeTest, []string{"read"}, nil, userID)
		require.NoError(t, err)
		require.NotEmpty(t, fullKey)

		req := httptest.NewRequest("GET", "/api/v1/tenants/"+tenantID.String()+"/api-keys", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		claims := &service.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: userID.String(),
			},
		}
		req = req.WithContext(SetClaims(req.Context(), claims))

		w := httptest.NewRecorder()
		apiKeyHandler.List(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		keys := resp["keys"].([]any)
		require.GreaterOrEqual(t, len(keys), 1)
	})

	// Test: Get API Key
	t.Run("GetAPIKey", func(t *testing.T) {
		key, fullKey, err := apiKeyService.CreateKey(ctx, tenantID, "Get Test Key", domain.APIKeyTypeTest, []string{"read"}, nil, userID)
		require.NoError(t, err)
		require.NotEmpty(t, fullKey)

		req := httptest.NewRequest("GET", "/api/v1/tenants/"+tenantID.String()+"/api-keys/"+key.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer test-token")

		claims := &service.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: userID.String(),
			},
		}
		req = req.WithContext(SetClaims(req.Context(), claims))

		w := httptest.NewRecorder()
		apiKeyHandler.Get(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp domain.APIKey
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		require.Equal(t, key.ID, resp.ID)
		require.Equal(t, "Get Test Key", resp.Name)
	})

	// Test: Revoke API Key
	t.Run("RevokeAPIKey", func(t *testing.T) {
		key, fullKey, err := apiKeyService.CreateKey(ctx, tenantID, "Revoke Test Key", domain.APIKeyTypeTest, []string{"read"}, nil, userID)
		require.NoError(t, err)
		require.NotEmpty(t, fullKey)

		req := httptest.NewRequest("POST", "/api/v1/tenants/"+tenantID.String()+"/api-keys/"+key.ID.String()+"/revoke", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		claims := &service.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: userID.String(),
			},
		}
		req = req.WithContext(SetClaims(req.Context(), claims))

		w := httptest.NewRecorder()
		apiKeyHandler.Revoke(w, req)

		require.Equal(t, http.StatusNoContent, w.Code)

		// Verify key is revoked
		retrievedKey, err := apiKeyService.GetKey(ctx, key.ID, tenantID)
		require.NoError(t, err)
		require.Equal(t, domain.APIKeyStatusRevoked, retrievedKey.Status)
	})

	// Test: Delete API Key
	t.Run("DeleteAPIKey", func(t *testing.T) {
		key, fullKey, err := apiKeyService.CreateKey(ctx, tenantID, "Delete Test Key", domain.APIKeyTypeTest, []string{"read"}, nil, userID)
		require.NoError(t, err)
		require.NotEmpty(t, fullKey)

		req := httptest.NewRequest("DELETE", "/api/v1/tenants/"+tenantID.String()+"/api-keys/"+key.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer test-token")

		claims := &service.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: userID.String(),
			},
		}
		req = req.WithContext(SetClaims(req.Context(), claims))

		w := httptest.NewRecorder()
		apiKeyHandler.Delete(w, req)

		require.Equal(t, http.StatusNoContent, w.Code)

		// Verify key is deleted
		_, err = apiKeyService.GetKey(ctx, key.ID, tenantID)
		require.Error(t, err)
	})

	// Test: Validate API Key
	t.Run("ValidateAPIKey", func(t *testing.T) {
		key, fullKey, err := apiKeyService.CreateKey(ctx, tenantID, "Validate Test Key", domain.APIKeyTypeTest, []string{"read"}, nil, userID)
		require.NoError(t, err)
		require.NotEmpty(t, fullKey)

		// Valid key
		validatedKey, err := apiKeyService.ValidateKey(ctx, fullKey)
		require.NoError(t, err)
		require.Equal(t, key.ID, validatedKey.ID)
		require.True(t, validatedKey.IsActive())

		// Invalid key
		_, err = apiKeyService.ValidateKey(ctx, "sk_test_invalidkey")
		require.Error(t, err)
	})

	// Test: API Key with Expiration
	t.Run("APIKeyWithExpiration", func(t *testing.T) {
		expiresAt := time.Now().Add(24 * time.Hour)
		key, fullKey, err := apiKeyService.CreateKey(ctx, tenantID, "Expiring Key", domain.APIKeyTypeTest, []string{"read"}, &expiresAt, userID)
		require.NoError(t, err)
		require.NotEmpty(t, fullKey)

		require.NotNil(t, key.ExpiresAt)
		require.True(t, key.IsActive())

		// Test expired key
		pastExpiration := time.Now().Add(-1 * time.Hour)
		expiredKey, _, err := apiKeyService.CreateKey(ctx, tenantID, "Expired Key", domain.APIKeyTypeTest, []string{"read"}, &pastExpiration, userID)
		require.NoError(t, err)

		require.False(t, expiredKey.IsActive())
	})
}
