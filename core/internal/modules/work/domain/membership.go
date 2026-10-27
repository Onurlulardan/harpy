package domain

import "errors"

var ErrLastUserAdmin = errors.New("work board requires an active user administrator")

type Membership struct {
	ID, BoardID ID
	Principal   Principal
	Role        BoardRole
	Active      bool
	RowVersion  int64
}

func (membership Membership) Valid() bool {
	return membership.ID.Valid() && membership.BoardID.Valid() &&
		membership.Principal.Valid() && membership.Role.Valid() && membership.RowVersion > 0
}

func EffectiveRole(memberships []Membership, principals []Principal) BoardRole {
	roles := make([]BoardRole, 0, len(memberships))
	for _, membership := range memberships {
		if !membership.Active || !membership.Valid() {
			continue
		}
		for _, principal := range principals {
			if membership.Principal == principal {
				roles = append(roles, membership.Role)
				break
			}
		}
	}
	return HighestRole(roles...)
}

func ValidateAdminRemoval(memberships []Membership, target ID) error {
	activeUserAdmins := 0
	removesUserAdmin := false
	for _, membership := range memberships {
		if membership.Active && membership.Principal.Kind == PrincipalUser &&
			membership.Role == RoleAdmin {
			activeUserAdmins++
			if membership.ID == target {
				removesUserAdmin = true
			}
		}
	}
	if removesUserAdmin && activeUserAdmins <= 1 {
		return ErrLastUserAdmin
	}
	return nil
}

func HasViewerAccess(memberships []Membership, principal Principal) bool {
	for _, membership := range memberships {
		if membership.Active && membership.Principal == principal && membership.Role.Rank() >= RoleViewer.Rank() {
			return true
		}
	}
	return false
}
