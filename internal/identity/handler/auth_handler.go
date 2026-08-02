// Package handler provides HTTP handlers for the identity module.
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/medaminerjb/saas-kit/internal/identity/domain"
	"github.com/medaminerjb/saas-kit/internal/identity/service"
)

// AuthHandler handles authentication HTTP endpoints.
type AuthHandler struct {
	identity *service.IdentityManager
	logger   *slog.Logger
}

// NewAuthHandler creates a new authentication handler.
func NewAuthHandler(identity *service.IdentityManager, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{identity: identity, logger: logger}
}

// Routes registers auth routes on the given router.
func (h *AuthHandler) Routes(r chi.Router) {
	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/refresh", h.RefreshTokens)
	r.Post("/auth/forgot-password", h.ForgotPassword)
	r.Post("/auth/reset-password", h.ResetPassword)
	r.Post("/auth/verify-email", h.VerifyEmail)
}

// ProtectedRoutes registers auth routes that require authentication.
func (h *AuthHandler) ProtectedRoutes(r chi.Router) {
	r.Post("/auth/logout", h.Logout)
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, tokens, err := h.identity.Register(r.Context(), service.RegisterInput{
		Email:    strings.TrimSpace(req.Email),
		Password: req.Password,
		Name:     strings.TrimSpace(req.Name),
	})
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, tokens, err := h.identity.Login(r.Context(), service.LoginInput{
		Email:     strings.TrimSpace(req.Email),
		Password:  req.Password,
		UserAgent: r.UserAgent(),
		IPAddress: extractIP(r),
	})
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokens handles POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tokens, err := h.identity.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	sessionID, err := uuid.Parse(claims.SessionID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session")
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user")
		return
	}

	if err := h.identity.Logout(r.Context(), sessionID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "logout failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

// ForgotPassword handles POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Always return success to prevent email enumeration
	_ = h.identity.Auth.RequestPasswordReset(r.Context(), strings.TrimSpace(req.Email))

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "if an account with that email exists, a reset link has been sent",
	})
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ResetPassword handles POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.identity.Auth.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		h.handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "password reset successfully"})
}

type verifyEmailRequest struct {
	Token string `json:"token"`
}

// VerifyEmail handles POST /api/v1/auth/verify-email
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.identity.Auth.VerifyEmail(r.Context(), req.Token); err != nil {
		h.handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "email verified"})
}

func (h *AuthHandler) handleAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrUserAlreadyExists):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, domain.ErrAccountDisabled):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, domain.ErrAccountLocked):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, domain.ErrInvalidToken):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, domain.ErrSessionExpired):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, domain.ErrInvalidEmail):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrPasswordTooShort),
		errors.Is(err, domain.ErrPasswordTooLong),
		errors.Is(err, domain.ErrPasswordRequired):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.logger.Error("auth error", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}
