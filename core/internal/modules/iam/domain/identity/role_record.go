package identity

import "time"

type Role struct {
	ID             IdentityID
	ScopeKind      RoleScopeKind
	ScopeID        IdentityID
	OrganizationID IdentityID
	Name           string
	Description    string
	Status         RoleStatus
	Protected      bool
	Default        bool
	Revision       int64
	Profile        ProfileDocument
	CreatedAt      time.Time
	UpdatedAt      time.Time
	RetiredAt      time.Time
}

func (role Role) Validate() error {
	if role.ID == (IdentityID{}) || !ValidRoleScope(
		role.ScopeKind, role.ScopeID, role.OrganizationID,
	) || !ValidDisplayName(role.Name) || len(role.Description) > 2000 ||
		!role.Status.Valid() || role.Revision < 1 ||
		role.Profile.Validate(ProfileRole) != nil || role.CreatedAt.IsZero() ||
		role.UpdatedAt.Before(role.CreatedAt) ||
		(role.Status == RoleRetired) != !role.RetiredAt.IsZero() {
		return ErrInvalidRecord
	}
	return nil
}
