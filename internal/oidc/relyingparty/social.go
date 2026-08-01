// Package relyingparty implements OAuth2/OIDC social login (SaaSKit acts as a relying party).
// Supported providers: Google, GitHub. Extensible to any OAuth2/OIDC provider.
package relyingparty

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/saaskit/saaskit/internal/identity/domain"
	"github.com/saaskit/saaskit/internal/identity/repository"
	"github.com/saaskit/saaskit/internal/identity/service"
	"github.com/saaskit/saaskit/internal/platform/events"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// Provider represents a configured OAuth2 social login provider.
type Provider struct {
	Name        string // "google", "github"
	Config      *oauth2.Config
	UserInfoURL string
	MapUser     func(data map[string]any) (*UserInfo, error)
}

// UserInfo is the normalized user profile extracted from an OAuth2 provider.
type UserInfo struct {
	ExternalID string
	Email      string
	Name       string
	AvatarURL  string
	Verified   bool
}

// Handler manages OAuth2 social login flows.
type Handler struct {
	providers       map[string]*Provider
	userRepo        repository.UserRepository
	identityManager *service.IdentityManager
	publisher       events.Publisher
	baseURL         string
	logger          *slog.Logger
}

// Config holds the handler configuration.
type Config struct {
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string
	BaseURL            string
	UserRepo           repository.UserRepository
	IdentityManager    *service.IdentityManager
	Publisher          events.Publisher
	Logger             *slog.Logger
}

// NewHandler creates a new social login handler.
func NewHandler(cfg Config) *Handler {
	h := &Handler{
		providers:       make(map[string]*Provider),
		userRepo:        cfg.UserRepo,
		identityManager: cfg.IdentityManager,
		publisher:       cfg.Publisher,
		baseURL:         cfg.BaseURL,
		logger:          cfg.Logger,
	}

	// Register Google
	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		h.providers["google"] = &Provider{
			Name: "google",
			Config: &oauth2.Config{
				ClientID:     cfg.GoogleClientID,
				ClientSecret: cfg.GoogleClientSecret,
				Endpoint:     google.Endpoint,
				RedirectURL:  cfg.BaseURL + "/oauth2/google/callback",
				Scopes:       []string{"openid", "email", "profile"},
			},
			UserInfoURL: "https://www.googleapis.com/oauth2/v3/userinfo",
			MapUser:     mapGoogleUser,
		}
	}

	// Register GitHub
	if cfg.GitHubClientID != "" && cfg.GitHubClientSecret != "" {
		h.providers["github"] = &Provider{
			Name: "github",
			Config: &oauth2.Config{
				ClientID:     cfg.GitHubClientID,
				ClientSecret: cfg.GitHubClientSecret,
				Endpoint:     github.Endpoint,
				RedirectURL:  cfg.BaseURL + "/oauth2/github/callback",
				Scopes:       []string{"user:email", "read:user"},
			},
			UserInfoURL: "https://api.github.com/user",
			MapUser:     mapGitHubUser,
		}
	}

	return h
}

// Routes registers social login routes.
func (h *Handler) Routes(r chi.Router) {
	r.Get("/{provider}/login", h.Login)
	r.Get("/{provider}/callback", h.Callback)
}

// Login redirects the user to the OAuth2 provider's authorization endpoint.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")
	provider, ok := h.providers[providerName]
	if !ok {
		http.Error(w, fmt.Sprintf("provider %q not configured", providerName), http.StatusNotFound)
		return
	}

	// Generate CSRF state token
	state := uuid.NewString()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	})

	url := provider.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Callback handles the OAuth2 callback after the user authorizes with the provider.
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")
	provider, ok := h.providers[providerName]
	if !ok {
		http.Error(w, "provider not configured", http.StatusNotFound)
		return
	}

	// Verify state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "invalid state parameter", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing authorization code", http.StatusBadRequest)
		return
	}

	token, err := provider.Config.Exchange(r.Context(), code)
	if err != nil {
		h.logger.Error("token exchange failed", slog.String("provider", providerName), slog.String("error", err.Error()))
		http.Error(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	// Fetch user info
	userInfo, err := h.fetchUserInfo(r.Context(), provider, token)
	if err != nil {
		h.logger.Error("user info fetch failed", slog.String("provider", providerName), slog.String("error", err.Error()))
		http.Error(w, "failed to get user profile", http.StatusInternalServerError)
		return
	}

	// Find or create user
	user, authTokens, err := h.findOrCreateUser(r.Context(), providerName, userInfo)
	if err != nil {
		h.logger.Error("user creation failed", slog.String("error", err.Error()))
		http.Error(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	_ = h.publisher.Publish(r.Context(), events.Event{
		Type:      "user.social_login",
		ActorID:   &user.ID,
		TargetID:  &user.ID,
		Payload:   map[string]string{"provider": providerName},
		Timestamp: time.Now(),
	})

	// Return tokens as JSON (frontend SPA flow)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"user":   user,
		"tokens": authTokens,
	})
}

func (h *Handler) fetchUserInfo(ctx context.Context, provider *Provider, token *oauth2.Token) (*UserInfo, error) {
	client := provider.Config.Client(ctx, token)
	resp, err := client.Get(provider.UserInfoURL)
	if err != nil {
		return nil, fmt.Errorf("fetching user info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return provider.MapUser(data)
}

func (h *Handler) findOrCreateUser(ctx context.Context, providerName string, info *UserInfo) (*domain.User, *service.AuthTokens, error) {
	// Try to find existing user by email
	existing, err := h.userRepo.GetByEmail(ctx, info.Email, nil)
	if err == nil && existing != nil {
		// User exists — create session for them (no password needed, already authed via provider)
		tokens, err := h.identityManager.Auth.CreateSessionForUser(ctx, existing)
		if err != nil {
			return nil, nil, err
		}
		return existing, tokens, nil
	}

	// Create new user with password (social login users get a random placeholder)
	user, tokens, err := h.identityManager.Register(ctx, service.RegisterInput{
		Email:    info.Email,
		Password: uuid.NewString() + "!Aa1", // Random strong password — user will use social login
		Name:     info.Name,
	})
	if err != nil {
		return nil, nil, err
	}

	// Set email as verified and avatar
	_ = h.userRepo.SetEmailVerified(ctx, user.ID)
	user.EmailVerified = true
	user.Status = domain.UserStatusActive
	if info.AvatarURL != "" {
		user.AvatarURL = &info.AvatarURL
		_ = h.userRepo.Update(ctx, user)
	}

	return user, tokens, nil
}

// ───────────────────────────────────────────────────────
// Provider-specific user mapping
// ───────────────────────────────────────────────────────

func mapGoogleUser(data map[string]any) (*UserInfo, error) {
	return &UserInfo{
		ExternalID: getString(data, "sub"),
		Email:      getString(data, "email"),
		Name:       getString(data, "name"),
		AvatarURL:  getString(data, "picture"),
		Verified:   getBool(data, "email_verified"),
	}, nil
}

func mapGitHubUser(data map[string]any) (*UserInfo, error) {
	id := ""
	if v, ok := data["id"]; ok {
		id = fmt.Sprintf("%v", v)
	}
	email := getString(data, "email")
	name := getString(data, "name")
	if name == "" {
		name = getString(data, "login")
	}
	return &UserInfo{
		ExternalID: id,
		Email:      email,
		Name:       name,
		AvatarURL:  getString(data, "avatar_url"),
		Verified:   true, // GitHub doesn't return verification status in /user
	}, nil
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
