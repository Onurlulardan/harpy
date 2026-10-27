package identity

type RoleScopeKind string

const (
	RoleScopeTenant       RoleScopeKind = "tenant"
	RoleScopeOrganization RoleScopeKind = "organization"
)

func (kind RoleScopeKind) Valid() bool {
	return kind == RoleScopeTenant || kind == RoleScopeOrganization
}

func ValidRoleScope(kind RoleScopeKind, scopeID, organizationID IdentityID) bool {
	if !kind.Valid() || scopeID == (IdentityID{}) {
		return false
	}
	return kind == RoleScopeTenant && organizationID == (IdentityID{}) ||
		kind == RoleScopeOrganization && organizationID == scopeID
}
