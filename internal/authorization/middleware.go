package authorization

import (
	"context"
	"encoding/json"
	"net/http"

	tenantdomain "github.com/medaminerjb/saas-kit/internal/tenant/domain"
)

// RoleGetter extracts the caller's tenant role from request context.
type RoleGetter func(ctx context.Context) (tenantdomain.MemberRole, bool)

// RequirePermission returns middleware that enforces a tenant-scoped permission.
// The RoleGetter must be wired to extract the role set by tenant membership middleware.
func RequirePermission(perm Permission, getRole RoleGetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := getRole(r.Context())
			if !ok {
				writeForbidden(w, "tenant context required")
				return
			}
			if !HasPermission(role, perm) {
				writeForbidden(w, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeForbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":   http.StatusText(http.StatusForbidden),
		"message": message,
	})
}
