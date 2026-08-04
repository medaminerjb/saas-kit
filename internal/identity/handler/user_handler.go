package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/medaminerjb/saas-kit/internal/identity/service"
)

// UserHandler handles user profile HTTP endpoints.
type UserHandler struct {
	identity *service.IdentityManager
	logger   *slog.Logger
}

// NewUserHandler creates a new user handler.
func NewUserHandler(identity *service.IdentityManager, logger *slog.Logger) *UserHandler {
	return &UserHandler{identity: identity, logger: logger}
}

// Routes registers user routes (all require authentication).
func (h *UserHandler) Routes(r chi.Router) {
	r.Get("/users/me", h.GetMe)
	r.Patch("/users/me", h.UpdateMe)
	r.Get("/users/me/sessions", h.ListSessions)
	r.Delete("/users/me/sessions/{sessionID}", h.RevokeSession)
	r.Get("/users/me/metadata", h.GetMetadata)
	r.Patch("/users/me/metadata", h.UpdateMetadata)
}

// GetMe handles GET /api/v1/users/me
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.identity.GetCurrentUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

type updateMeRequest struct {
	Name      *string `json:"name,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// UpdateMe handles PATCH /api/v1/users/me
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req updateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.identity.UpdateProfile(r.Context(), userID, service.UpdateProfileInput{
		Name:      req.Name,
		AvatarURL: req.AvatarURL,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "update failed")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// ListSessions handles GET /api/v1/users/me/sessions
func (h *UserHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Stub: will be implemented when session listing is wired through the IdentityManager
	writeJSON(w, http.StatusOK, map[string]any{"sessions": []any{}})
}

// RevokeSession handles DELETE /api/v1/users/me/sessions/{sessionID}
func (h *UserHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if err := h.identity.Logout(r.Context(), sessionID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "revoke failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "session revoked"})
}

// GetMetadata handles GET /api/v1/users/me/metadata
func (h *UserHandler) GetMetadata(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.identity.GetCurrentUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"metadata_public": user.MetadataPublic,
	})
}

type updateMetadataRequest struct {
	MetadataPublic  map[string]interface{} `json:"metadata_public,omitempty"`
	MetadataPrivate map[string]interface{} `json:"metadata_private,omitempty"`
}

// UpdateMetadata handles PATCH /api/v1/users/me/metadata
func (h *UserHandler) UpdateMetadata(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req updateMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.identity.UpdateUserMetadata(r.Context(), userID, service.UpdateMetadataInput{
		MetadataPublic:  req.MetadataPublic,
		MetadataPrivate: req.MetadataPrivate,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "update failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"metadata_public": user.MetadataPublic,
	})
}
