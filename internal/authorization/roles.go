package authorization

import (
	tenantdomain "github.com/medaminerjb/saas-kit/internal/tenant/domain"
)

var rolePermissions = map[tenantdomain.MemberRole][]Permission{
	tenantdomain.RoleOwner: {
		PermTenantRead, PermTenantUpdate, PermTenantDelete,
		PermMembersList, PermMembersInvite, PermMembersRemove, PermMembersUpdateRole,
		PermTenantMetadataRead, PermTenantMetadataWrite,
		PermUserMetadataRead, PermUserMetadataWrite,
	},
	tenantdomain.RoleAdmin: {
		PermTenantRead, PermTenantUpdate,
		PermMembersList, PermMembersInvite, PermMembersRemove,
		PermTenantMetadataRead, PermTenantMetadataWrite,
		PermUserMetadataRead, PermUserMetadataWrite,
	},
	tenantdomain.RoleManager: {
		PermTenantRead,
		PermMembersList, PermMembersInvite,
		PermTenantMetadataRead, PermTenantMetadataWrite,
		PermUserMetadataRead,
	},
	tenantdomain.RoleMember: {
		PermTenantRead,
		PermMembersList,
		PermTenantMetadataRead,
		PermUserMetadataRead,
	},
	tenantdomain.RoleViewer: {
		PermTenantRead,
		PermMembersList,
		PermTenantMetadataRead,
		PermUserMetadataRead,
	},
}

// HasPermission returns true if the given role grants the permission.
func HasPermission(role tenantdomain.MemberRole, perm Permission) bool {
	perms, ok := rolePermissions[role]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}

// PermissionsForRole returns all permissions granted to a role.
func PermissionsForRole(role tenantdomain.MemberRole) []Permission {
	return rolePermissions[role]
}

// ValidRole returns true if role is a recognized member role.
func ValidRole(role tenantdomain.MemberRole) bool {
	_, ok := rolePermissions[role]
	return ok
}
