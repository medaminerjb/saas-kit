package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medaminerjb/saas-kit/internal/platform/database"
	"github.com/medaminerjb/saas-kit/internal/platform/events"

	"github.com/medaminerjb/saas-kit/internal/identity/service"
	tenantservice "github.com/medaminerjb/saas-kit/internal/tenant/service"
)

// RouterConfig holds optional components for the main router.
type RouterConfig struct {
	Identity *service.IdentityManager
	Pool     *pgxpool.Pool
	Logger   *slog.Logger

	// OIDCProvider is the mounted OIDC provider handler (nil to disable).
	OIDCProvider http.Handler
	// SocialLogin is the mounted social login handler (nil to disable).
	SocialLogin http.Handler
	// TenantService is the tenant management service.
	TenantService *tenantservice.TenantService
}

// NewRouter creates the main HTTP router with all middleware and routes.
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()
	identity := cfg.Identity
	pool := cfg.Pool
	logger := cfg.Logger

	// ─── Global middleware ────────────────────────────
	r.Use(chimiddleware.RequestID)
	r.Use(clientInfoMiddleware)
	r.Use(SecurityHeadersMiddleware)
	
	// IP-based rate limiting: 20 req/sec per IP, burst of 40
	limiter := NewRateLimiter(20.0, 40.0)
	r.Use(limiter.Limit)

	r.Use(slogMiddleware(logger))
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))
	r.Use(corsMiddleware)

	// ─── Health endpoints ────────────────────────────
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		if err := database.HealthCheck(r.Context(), pool); err != nil {
			writeError(w, http.StatusServiceUnavailable, "database not ready")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	// ─── Auth handlers ───────────────────────────────
	authHandler := NewAuthHandler(identity, logger)
	userHandler := NewUserHandler(identity, logger)
	authMiddleware := NewAuthMiddleware(identity.Tokens, logger)

	// Public API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public (no auth required)
		authHandler.Routes(r)

		// Protected (JWT required)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Handler)
			authHandler.ProtectedRoutes(r)
			userHandler.Routes(r)
			if cfg.TenantService != nil {
				tenantHandler := NewTenantHandler(cfg.TenantService, logger)
				tenantHandler.Routes(r)
			}
		})
	})

	// ─── OIDC Provider ───────────────────────────────
	if cfg.OIDCProvider != nil {
		r.Mount("/", cfg.OIDCProvider)
	} else {
		r.Get("/keys", func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{"keys": []any{}})
		})
	}

	// ─── OAuth2 social login ─────────────────────────
	if cfg.SocialLogin != nil {
		r.Mount("/oauth2", cfg.SocialLogin)
	} else {
		r.Route("/oauth2", func(r chi.Router) {
			r.Get("/{provider}/login", func(w http.ResponseWriter, _ *http.Request) {
				writeError(w, http.StatusNotImplemented, "social login not configured")
			})
			r.Get("/{provider}/callback", func(w http.ResponseWriter, _ *http.Request) {
				writeError(w, http.StatusNotImplemented, "social login not configured")
			})
		})
	}

	// ─── OIDC Discovery ──────────────────────────────
	// The zitadel/oidc provider serves its own discovery document at /.well-known/openid-configuration
	// This fallback is only used when the OIDC provider is not mounted.
	if cfg.OIDCProvider == nil {
		r.Get("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			baseURL := r.Header.Get("X-Forwarded-Proto") + "://" + r.Host
			if baseURL == "://" {
				baseURL = "http://" + r.Host
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"issuer":                                baseURL,
				"authorization_endpoint":                baseURL + "/authorize",
				"token_endpoint":                        baseURL + "/oauth/token",
				"userinfo_endpoint":                     baseURL + "/userinfo",
				"jwks_uri":                              baseURL + "/keys",
				"revocation_endpoint":                   baseURL + "/revoke",
				"introspection_endpoint":                baseURL + "/oauth/introspect",
				"response_types_supported":              []string{"code"},
				"subject_types_supported":               []string{"public"},
				"id_token_signing_alg_values_supported": []string{"RS256", "ES256", "EdDSA"},
				"scopes_supported":                      []string{"openid", "profile", "email"},
				"token_endpoint_auth_methods_supported": []string{"client_secret_basic", "client_secret_post", "none"},
				"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
				"code_challenge_methods_supported":      []string{"S256"},
			})
		})
	}

	return r
}

// slogMiddleware logs HTTP requests using slog.
func slogMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			logger.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Duration("duration", time.Since(start)),
				slog.String("request_id", chimiddleware.GetReqID(r.Context())),
			)
		})
	}
}

// corsMiddleware adds CORS headers. In production, this should be configured per-deployment.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// clientInfoMiddleware injects the client IP and User Agent into request context.
func clientInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)
		userAgent := r.UserAgent()
		ctx := events.WithClientInfo(r.Context(), ip, userAgent)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

