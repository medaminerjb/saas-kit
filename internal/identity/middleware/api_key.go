package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/medaminerjb/saas-kit/internal/identity/domain"
	"github.com/medaminerjb/saas-kit/internal/identity/service"
)

// APIKeyContextKey is the context key for storing the validated API key.
type APIKeyContextKey struct{}

// TenantIDContextKey is the context key for storing the tenant ID from API key auth.
type TenantIDContextKey struct{}

// APIKeyAuth middleware validates API keys and extracts tenant context.
func APIKeyAuth(apiKeyService *service.APIKeyService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Expect format: "Bearer sk_live_..." or "Bearer sk_test_..."
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			fullKey := strings.TrimPrefix(authHeader, "Bearer ")
			if fullKey == "" {
				http.Error(w, "API key required", http.StatusUnauthorized)
				return
			}

			// Validate the API key
			key, err := apiKeyService.ValidateKey(r.Context(), fullKey)
			if err != nil {
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}

			// Add the API key and tenant ID to the context
			ctx := context.WithValue(r.Context(), APIKeyContextKey{}, key)
			ctx = context.WithValue(ctx, TenantIDContextKey{}, key.TenantID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAPIKeyFromContext retrieves the API key from the request context.
func GetAPIKeyFromContext(ctx context.Context) (*domain.APIKey, bool) {
	key, ok := ctx.Value(APIKeyContextKey{}).(*domain.APIKey)
	return key, ok
}

// RequireScope middleware checks if the API key has the required scope.
func RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key, ok := GetAPIKeyFromContext(r.Context())
			if !ok {
				http.Error(w, "API key not found in context", http.StatusUnauthorized)
				return
			}

			if !key.HasScope(scope) {
				http.Error(w, "Insufficient scope", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
