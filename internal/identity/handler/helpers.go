package handler

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/medaminerjb/saas-kit/internal/identity/service"
)

type contextKey string

const claimsKey contextKey = "claims"

// SetClaims stores JWT claims in the request context.
func SetClaims(ctx context.Context, claims *service.JWTClaims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

// GetClaims retrieves JWT claims from the request context.
func GetClaims(ctx context.Context) *service.JWTClaims {
	claims, _ := ctx.Value(claimsKey).(*service.JWTClaims)
	return claims
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// writeError writes an error JSON response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"error":   http.StatusText(status),
		"message": message,
	})
}

// extractIP extracts the client IP from the request, respecting X-Forwarded-For.
func extractIP(r *http.Request) string {
	// Check X-Forwarded-For first (behind reverse proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// requireTenantMembership is a middleware that verifies the authenticated user belongs to the requested tenant.
// If valid, it injects tenant_id and role into the request context.
// This is a simplified version that requires the tenant service to be passed in.
// For now, we'll create a placeholder that can be wired up later.
func requireTenantMembership(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement tenant membership check
		// This requires access to tenantService which needs to be passed in
		// For now, we'll just pass through - this should be wired up properly
		next.ServeHTTP(w, r)
	})
}
