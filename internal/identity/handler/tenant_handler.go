package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/medaminerjb/saas-kit/internal/authorization"
	"github.com/medaminerjb/saas-kit/internal/tenant/domain"
	tenantservice "github.com/medaminerjb/saas-kit/internal/tenant/service"
)

type tenantContextKey string

const (
	tenantIDKey   tenantContextKey = "tenant_id"
	tenantRoleKey tenantContextKey = "tenant_role"
)

// TenantHandler handles tenant/organization administration API endpoints.
type TenantHandler struct {
	tenantService *tenantservice.TenantService
	logger        *slog.Logger
}

// NewTenantHandler creates a new tenant HTTP handler.
func NewTenantHandler(tenantService *tenantservice.TenantService, logger *slog.Logger) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
		logger:        logger,
	}
}

// Routes registers tenant-related routes.
func (h *TenantHandler) Routes(r chi.Router) {
	r.Post("/tenants", h.Create)
	r.Get("/tenants", h.List)
	r.Post("/tenants/switch", h.Switch)
	r.Post("/tenants/invitations/accept", h.AcceptInvitation)

	r.Route("/tenants/{tenantID}", func(r chi.Router) {
		r.Use(h.RequireTenantMembership)
		r.Get("/", h.Get)
		r.With(authorization.RequirePermission(authorization.PermTenantUpdate, GetTenantRole)).Patch("/", h.Update)
		r.Get("/members", h.ListMembers)
		r.With(authorization.RequirePermission(authorization.PermMembersInvite, GetTenantRole)).Post("/members", h.InviteMember)
		r.Delete("/members/{userID}", h.RemoveMember)
	})
}

type createTenantRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
}

func (h *TenantHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req createTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	t, err := h.tenantService.CreateTenant(r.Context(), req.Name, req.Slug, userID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, t)
}

func (h *TenantHandler) List(w http.ResponseWriter, r *http.Request) {
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

	tenants, roles, err := h.tenantService.ListTenantsForUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list organizations")
		return
	}

	type tenantResponseItem struct {
		*domain.Tenant
		Role domain.MemberRole `json:"role"`
	}

	items := make([]tenantResponseItem, len(tenants))
	for i, t := range tenants {
		items[i] = tenantResponseItem{
			Tenant: t,
			Role:   roles[i],
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"tenants": items})
}

type switchTenantRequest struct {
	TenantID string `json:"tenant_id"`
}

func (h *TenantHandler) Switch(w http.ResponseWriter, r *http.Request) {
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

	var req switchTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tID, err := uuid.Parse(req.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	if err := h.tenantService.SwitchActiveTenant(r.Context(), userID, tID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "active tenant switched"})
}

type acceptInvitationRequest struct {
	Token string `json:"token"`
}

func (h *TenantHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
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

	var req acceptInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Token == "" {
		writeError(w, http.StatusBadRequest, "token is required")
		return
	}

	if err := h.tenantService.AcceptInvitation(r.Context(), req.Token, userID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "invitation accepted"})
}

func (h *TenantHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusNotFound, "tenant context missing")
		return
	}

	t, err := h.tenantService.GetTenant(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}

	writeJSON(w, http.StatusOK, t)
}

type updateTenantRequest struct {
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

func (h *TenantHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusNotFound, "tenant context missing")
		return
	}

	var req updateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	t, err := h.tenantService.UpdateTenant(r.Context(), tenantID, req.Name, req.Slug)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, t)
}

func (h *TenantHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusNotFound, "tenant context missing")
		return
	}

	members, err := h.tenantService.ListMembers(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list members")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"members": members})
}

type inviteMemberRequest struct {
	Email string            `json:"email"`
	Role  domain.MemberRole `json:"role"`
}

func (h *TenantHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusNotFound, "tenant context missing")
		return
	}

	var req inviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	invite, rawToken, err := h.tenantService.InviteMember(r.Context(), tenantID, req.Email, req.Role)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":         invite.ID,
		"tenant_id":  invite.TenantID,
		"email":      invite.Email,
		"role":       invite.Role,
		"expires_at": invite.ExpiresAt,
		"token":      rawToken, // return raw token for development/testing convenience
	})
}

func (h *TenantHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := GetTenantID(r.Context())
	role, ok2 := GetTenantRole(r.Context())
	if !ok || !ok2 {
		writeError(w, http.StatusNotFound, "tenant context missing")
		return
	}

	targetUserIDStr := chi.URLParam(r, "userID")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// Users can remove themselves (leaving the organization) regardless of role
	claims := GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	currentUserID, _ := uuid.Parse(claims.Subject)

	isSelf := currentUserID == targetUserID
	if !isSelf && !authorization.HasPermission(role, authorization.PermMembersRemove) {
		writeError(w, http.StatusForbidden, "forbidden: insufficient permissions")
		return
	}

	if err := h.tenantService.RemoveMember(r.Context(), tenantID, targetUserID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "member removed successfully"})
}

// RequireTenantMembership is a middleware that verifies the authenticated user belongs to the requested tenant.
// If valid, it injects tenant_id and role into the request context.
func (h *TenantHandler) RequireTenantMembership(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		tenantIDStr := chi.URLParam(r, "tenantID")
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid tenant ID")
			return
		}

		// Retrieve membership details from repository directly (or through service)
		// We can get member details from the tenantService's repository
		members, err := h.tenantService.ListMembers(r.Context(), tenantID)
		if err != nil {
			writeError(w, http.StatusNotFound, "organization not found")
			return
		}

		var memberRole domain.MemberRole
		isMember := false
		for _, m := range members {
			if m.UserID == userID {
				isMember = true
				memberRole = m.Role
				break
			}
		}

		if !isMember {
			writeError(w, http.StatusForbidden, "forbidden: not a member of this organization")
			return
		}

		// Inject tenant details into context
		ctx := context.WithValue(r.Context(), tenantIDKey, tenantID)
		ctx = context.WithValue(ctx, tenantRoleKey, memberRole)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetTenantID extracts the tenant_id from the context.
func GetTenantID(ctx context.Context) (uuid.UUID, bool) {
	val, ok := ctx.Value(tenantIDKey).(uuid.UUID)
	return val, ok
}

// GetTenantRole extracts the member role from the context.
func GetTenantRole(ctx context.Context) (domain.MemberRole, bool) {
	val, ok := ctx.Value(tenantRoleKey).(domain.MemberRole)
	return val, ok
}
