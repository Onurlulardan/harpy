package identity

import "bytes"

func validOrganizationOwnerReplacement(command UpdateOrganizationCommand) bool {
	if !command.ReplaceOwners {
		return len(command.OwnerIDs) == 0 && len(command.OwnerMembershipIDs) == 0
	}
	return validOrganizationOwnerSet(command.OwnerIDs, command.OwnerMembershipIDs, true)
}

func validOrganizationOwnerSet(
	ownerIDs []IdentityID,
	membershipIDs []IdentityID,
	required bool,
) bool {
	if len(ownerIDs) != len(membershipIDs) || len(ownerIDs) > 100 ||
		required && len(ownerIDs) == 0 {
		return false
	}
	memberships := make(map[IdentityID]struct{}, len(membershipIDs))
	for index, ownerID := range ownerIDs {
		membershipID := membershipIDs[index]
		if ownerID == (IdentityID{}) || membershipID == (IdentityID{}) {
			return false
		}
		if index > 0 && bytes.Compare(ownerIDs[index-1][:], ownerID[:]) >= 0 {
			return false
		}
		if _, duplicate := memberships[membershipID]; duplicate {
			return false
		}
		memberships[membershipID] = struct{}{}
	}
	return true
}
