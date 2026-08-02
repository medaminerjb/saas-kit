package handler

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

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

// clientLimiter holds rate limiting state for a single IP.
type clientLimiter struct {
	tokens     float64
	lastRefill time.Time
}

// RateLimiter manages rate limiting for HTTP endpoints.
type RateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientLimiter
	rate     float64 // tokens per second
	burst    float64 // maximum tokens
	cleanupD time.Duration
}

// NewRateLimiter creates a new RateLimiter.
// rate is the number of allowed requests per second.
// burst is the maximum burst size.
func NewRateLimiter(rate, burst float64) *RateLimiter {
	rl := &RateLimiter{
		clients:  make(map[string]*clientLimiter),
		rate:     rate,
		burst:    burst,
		cleanupD: 10 * time.Minute,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	limiter, exists := rl.clients[ip]
	if !exists {
		limiter = &clientLimiter{
			tokens:     rl.burst,
			lastRefill: now,
		}
		rl.clients[ip] = limiter
	}

	// Refill tokens based on time elapsed
	elapsed := now.Sub(limiter.lastRefill).Seconds()
	limiter.lastRefill = now
	limiter.tokens += elapsed * rl.rate
	if limiter.tokens > rl.burst {
		limiter.tokens = rl.burst
	}

	if limiter.tokens >= 1.0 {
		limiter.tokens -= 1.0
		return true
	}

	return false
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupD)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, limiter := range rl.clients {
			// If client has not made a request in 10 minutes, remove from map
			if now.Sub(limiter.lastRefill) > rl.cleanupD {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Limit returns a middleware that rate limits requests.
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)
		if !rl.Allow(ip) {
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware adds standard security headers to all HTTP responses.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'; object-src 'none';")

		// If HTTPS, set Strict-Transport-Security
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		}

		next.ServeHTTP(w, r)
	})
}
