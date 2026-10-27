package adoption

import "github.com/google/uuid"

func (state *validationState) validateMappedIdentityCollisions() {
	identifiers := make(map[uuid.UUID]string)
	modes := make(map[uuid.UUID]string)
	memberships := make(map[uuid.UUID]string)
	assignments := make(map[uuid.UUID]string)
	bindings := make(map[uuid.UUID]string)
	for _, mapping := range state.mapping.Identity.Users {
		key := mapping.UserID.String()
		state.claimMappedID(identifiers, mapping.EmailIdentifierID, key, "identifier_id_collision")
		state.claimMappedID(identifiers, mapping.PortalIdentifierID, key, "identifier_id_collision")
		state.claimMappedID(modes, mapping.OperatorModeID, key, "access_mode_id_collision")
		state.claimMappedID(modes, mapping.PortalModeID, key, "access_mode_id_collision")
		state.claimMappedID(memberships, mapping.MembershipID, key, "membership_id_collision")
		state.claimMappedID(bindings, mapping.PortalBindingID, key, "portal_binding_id_collision")
		for _, assignment := range mapping.Assignments {
			id := assignment.AssignmentID
			state.claimMappedID(assignments, &id, key, "assignment_id_collision")
		}
	}
}

func (state *validationState) claimMappedID(
	seen map[uuid.UUID]string, value *uuid.UUID, sourceKey, code string,
) {
	if value == nil {
		return
	}
	if *value == uuid.Nil {
		state.add(issue("mapping_identity_invalid", "mapping", sourceKey))
		return
	}
	if prior, duplicate := seen[*value]; duplicate && prior != sourceKey {
		state.add(issue(code, "mapping", value.String()))
	}
	seen[*value] = sourceKey
}
