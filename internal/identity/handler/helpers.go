package handler

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/saaskit/saaskit/internal/identity/service"
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
