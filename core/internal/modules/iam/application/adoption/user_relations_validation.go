package adoption

import "github.com/google/uuid"

func (state *validationState) validateUserRelations(user SourceUser, mapping UserMapping) {
	var organizationID uuid.UUID
	if user.OrganizationID != "" {
		parsed, err := uuid.Parse(user.OrganizationID)
		if err != nil || state.orgs[parsed].ID == uuid.Nil {
			state.add(issue("dangling_organization", "memberships", user.ID.String()))
		} else {
			organizationID = parsed
			if nilUUID(mapping.MembershipID) {
				state.add(issue("membership_mapping_missing", "memberships", user.ID.String()))
			}
		}
	} else if mapping.MembershipID != nil {
		state.add(issue("membership_mapping_unexpected", "memberships", user.ID.String()))
	}
	assigned := make(map[uuid.UUID]bool)
	for _, rawRoleID := range user.RoleIDs {
		roleID, err := uuid.Parse(rawRoleID)
		if err != nil {
			state.add(issue("malformed_role_id", "assignments", user.ID.String()))
			continue
		}
		role, found := state.roles[roleID]
		if !found {
			state.add(issue("dangling_role", "assignments", user.ID.String()+":"+roleID.String()))
			continue
		}
		if assigned[roleID] {
			state.add(issue("duplicate_role_assignment", "assignments", user.ID.String()+":"+roleID.String()))
		}
		assigned[roleID] = true
		if role.OrganizationID != "" && role.OrganizationID != organizationID.String() {
			state.add(issue("cross_organization_role", "assignments", user.ID.String()+":"+roleID.String()))
		}
	}
	mapped := make(map[uuid.UUID]bool)
	for _, assignment := range mapping.Assignments {
		if assignment.RoleID == uuid.Nil || assignment.AssignmentID == uuid.Nil || mapped[assignment.RoleID] {
			state.add(issue("assignment_mapping_invalid", "assignments", user.ID.String()))
		}
		mapped[assignment.RoleID] = true
	}
	for roleID := range assigned {
		if !mapped[roleID] {
			state.add(issue("assignment_mapping_missing", "assignments", user.ID.String()+":"+roleID.String()))
		}
	}
	for roleID := range mapped {
		if !assigned[roleID] {
			state.add(issue("assignment_mapping_unexpected", "assignments", user.ID.String()+":"+roleID.String()))
		}
	}
}
