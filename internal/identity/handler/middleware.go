package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/saaskit/saaskit/internal/identity/service"
)

// AuthMiddleware validates JWT access tokens and injects claims into the request context.
type AuthMiddleware struct {
	tokenService *service.TokenService
	logger       *slog.Logger
}

// NewAuthMiddleware creates a new JWT authentication middleware.
func NewAuthMiddleware(tokenService *service.TokenService, logger *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{tokenService: tokenService, logger: logger}
}

// Handler returns an http.Handler middleware that validates Bearer tokens.
func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			writeError(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		claims, err := m.tokenService.ValidateAccessToken(parts[1])
		if err != nil {
			m.logger.Debug("token validation failed", slog.String("error", err.Error()))
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := SetClaims(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
