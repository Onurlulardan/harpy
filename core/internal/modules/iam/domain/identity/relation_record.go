package identity

import "time"

type Membership struct {
	ID             IdentityID
	OrganizationID IdentityID
	UserID         IdentityID
	Kind           MembershipKind
	State          MembershipState
	Revision       int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
	TombstonedAt   time.Time
}

func (membership Membership) Validate() error {
	if membership.ID == (IdentityID{}) || membership.OrganizationID == (IdentityID{}) ||
		membership.UserID == (IdentityID{}) || !membership.Kind.Valid() ||
		!membership.State.Valid() || membership.Revision < 1 ||
		membership.CreatedAt.IsZero() ||
		membership.UpdatedAt.Before(membership.CreatedAt) ||
		(membership.State == MembershipTombstoned) !=
			!membership.TombstonedAt.IsZero() {
		return ErrInvalidRecord
	}
	return nil
}

type Assignment struct {
	ID             IdentityID
	UserID         IdentityID
	RoleID         IdentityID
	ScopeKind      RoleScopeKind
	ScopeID        IdentityID
	OrganizationID IdentityID
	MembershipID   IdentityID
	State          AssignmentState
	Revision       int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
	RevokedAt      time.Time
}

func (assignment Assignment) Validate() error {
	if assignment.ID == (IdentityID{}) || assignment.UserID == (IdentityID{}) ||
		assignment.RoleID == (IdentityID{}) || !ValidRoleScope(
		assignment.ScopeKind, assignment.ScopeID, assignment.OrganizationID,
	) || assignment.Revision < 1 || !assignment.State.Valid() ||
		assignment.CreatedAt.IsZero() ||
		assignment.UpdatedAt.Before(assignment.CreatedAt) ||
		(assignment.State == AssignmentRevoked) != !assignment.RevokedAt.IsZero() {
		return ErrInvalidRecord
	}
	if assignment.ScopeKind == RoleScopeTenant && assignment.MembershipID != (IdentityID{}) ||
		assignment.ScopeKind == RoleScopeOrganization && assignment.MembershipID == (IdentityID{}) {
		return ErrInvalidRecord
	}
	return nil
}
