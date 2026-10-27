package systemresource

import identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"

type User = identitydomain.User
type Role = identitydomain.Role
type Organization = identitydomain.Organization
type UserStatus = identitydomain.UserStatus
type OrganizationStatus = identitydomain.OrganizationStatus
type RoleScopeKind = identitydomain.RoleScopeKind
type MembershipKind = identitydomain.MembershipKind
type DirectoryProjection = identitydomain.DirectoryProjection

var NewDirectoryProjection = identitydomain.NewDirectoryProjection

const (
	UserSuspended  = identitydomain.UserSuspended
	UserDisabled   = identitydomain.UserDisabled
	UserTombstoned = identitydomain.UserTombstoned
	UserActive     = identitydomain.UserActive

	OrganizationTombstoned = identitydomain.OrganizationTombstoned
	MembershipMember       = identitydomain.MembershipMember
	RoleScopeTenant        = identitydomain.RoleScopeTenant
	RoleScopeOrganization  = identitydomain.RoleScopeOrganization
)
