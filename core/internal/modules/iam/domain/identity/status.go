package identity

type UserStatus string

const (
	UserPending    UserStatus = "pending"
	UserActive     UserStatus = "active"
	UserSuspended  UserStatus = "suspended"
	UserDisabled   UserStatus = "disabled"
	UserTombstoned UserStatus = "tombstoned"
)

func (status UserStatus) Valid() bool {
	switch status {
	case UserPending, UserActive, UserSuspended, UserDisabled, UserTombstoned:
		return true
	default:
		return false
	}
}

type OrganizationStatus string

const (
	OrganizationActive     OrganizationStatus = "active"
	OrganizationSuspended  OrganizationStatus = "suspended"
	OrganizationTombstoned OrganizationStatus = "tombstoned"
)

func (status OrganizationStatus) Valid() bool {
	return status == OrganizationActive || status == OrganizationSuspended ||
		status == OrganizationTombstoned
}

type RoleStatus string

const (
	RoleActive  RoleStatus = "active"
	RoleRetired RoleStatus = "retired"
)

func (status RoleStatus) Valid() bool {
	return status == RoleActive || status == RoleRetired
}
