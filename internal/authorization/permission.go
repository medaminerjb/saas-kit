// Package authorization provides tenant-scoped RBAC permission checks.
package authorization

// Permission represents a granular resource-action permission string.
type Permission string

const (
	PermTenantRead       Permission = "tenant.read"
	PermTenantUpdate     Permission = "tenant.update"
	PermTenantDelete     Permission = "tenant.delete"
	PermMembersList      Permission = "members.list"
	PermMembersInvite    Permission = "members.invite"
	PermMembersRemove    Permission = "members.remove"
	PermMembersUpdateRole Permission = "members.update_role"
)
