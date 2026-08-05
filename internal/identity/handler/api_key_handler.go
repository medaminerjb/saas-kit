package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/medaminerjb/saas-kit/internal/identity/domain"
	"github.com/medaminerjb/saas-kit/internal/identity/service"
)

// APIKeyHandler handles API key HTTP requests.
type APIKeyHandler struct {
	apiKeyService *service.APIKeyService
	logger        *slog.Logger
}

// NewAPIKeyHandler creates a new APIKeyHandler.
func NewAPIKeyHandler(apiKeyService *service.APIKeyService, logger *slog.Logger) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService: apiKeyService,
		logger:        logger,
	}
}

// RegisterRoutes registers API key routes.
func (h *APIKeyHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/tenants/{tenantID}/api-keys", func(r chi.Router) {
		r.Use(requireTenantMembership)
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Route("/{keyID}", func(r chi.Router) {
			r.Get("/", h.Get)
			r.Delete("/", h.Delete)
			r.Post("/revoke", h.Revoke)
		})
	})
}

// CreateRequest represents the request to create an API key.
type CreateRequest struct {
	Name      string            `json:"name"`
	Type      domain.APIKeyType `json:"type"`
	Scopes    []string          `json:"scopes"`
	ExpiresAt *time.Time        `json:"expires_at,omitempty"`
}

// CreateResponse represents the response after creating an API key.
type CreateResponse struct {
	Key *domain.APIKey `json:"key"`
	// The full API key (only shown once)
	FullKey string `json:"full_key"`
}

// Create creates a new API key.
func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, err := uuid.Parse(chi.URLParam(r, "tenantID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	claims := GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID in claims")
		return
	}

	key, fullKey, err := h.apiKeyService.CreateKey(r.Context(), tenantID, req.Name, req.Type, req.Scopes, req.ExpiresAt, userID)
	if err != nil {
		h.logger.Error("failed to create API key", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create API key")
		return
	}

	resp := CreateResponse{
		Key:     key,
		FullKey: fullKey,
	}

	writeJSON(w, http.StatusCreated, resp)
}

// List lists all API keys for a tenant.
func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, err := uuid.Parse(chi.URLParam(r, "tenantID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	keys, err := h.apiKeyService.ListKeys(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("failed to list API keys", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list API keys")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"keys": keys})
}

// Get retrieves a specific API key.
func (h *APIKeyHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID, err := uuid.Parse(chi.URLParam(r, "tenantID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	keyID, err := uuid.Parse(chi.URLParam(r, "keyID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid key ID")
		return
	}

	key, err := h.apiKeyService.GetKey(r.Context(), keyID, tenantID)
	if err != nil {
		h.logger.Error("failed to get API key", "error", err)
		writeError(w, http.StatusNotFound, "API key not found")
		return
	}

	writeJSON(w, http.StatusOK, key)
}

// Revoke revokes an API key.
func (h *APIKeyHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	tenantID, err := uuid.Parse(chi.URLParam(r, "tenantID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	keyID, err := uuid.Parse(chi.URLParam(r, "keyID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid key ID")
		return
	}

	claims := GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID in claims")
		return
	}

	if err := h.apiKeyService.RevokeKey(r.Context(), keyID, tenantID, userID); err != nil {
		h.logger.Error("failed to revoke API key", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to revoke API key")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete permanently deletes an API key.
func (h *APIKeyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID, err := uuid.Parse(chi.URLParam(r, "tenantID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	keyID, err := uuid.Parse(chi.URLParam(r, "keyID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid key ID")
		return
	}

	if err := h.apiKeyService.DeleteKey(r.Context(), keyID, tenantID); err != nil {
		h.logger.Error("failed to delete API key", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to delete API key")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
