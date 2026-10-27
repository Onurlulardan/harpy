package adoption

import (
	"strings"

	"github.com/google/uuid"
)

func (state *validationState) validateProtectedAdministrator() {
	identity := state.mapping.Identity
	if identity.ProtectedRoleID == uuid.Nil || identity.ProtectedUserID == uuid.Nil ||
		identity.ProtectedPermissionID == uuid.Nil || identity.ProtectedBindingID == uuid.Nil ||
		identity.ProtectedPolicyID == uuid.Nil {
		state.add(issue("protected_mapping_invalid", "protected_administrator", "mapping"))
		return
	}
	role := state.roles[identity.ProtectedRoleID]
	user := state.users[identity.ProtectedUserID]
	if role.ID == uuid.Nil || user.ID == uuid.Nil ||
		!strings.EqualFold(strings.TrimSpace(user.Status), "ACTIVE") {
		state.add(issue("protected_identity_missing", "protected_administrator", identity.ProtectedUserID.String()))
		return
	}
	assigned := false
	for _, rawRoleID := range user.RoleIDs {
		roleID, err := uuid.Parse(rawRoleID)
		assigned = assigned || err == nil && roleID == role.ID
	}
	if !assigned {
		state.add(issue("protected_assignment_missing", "protected_administrator", user.ID.String()))
	}
	if !state.hasProtectedWildcard(role.ID) {
		state.add(issue("protected_wildcard_missing", "protected_administrator", role.ID.String()))
	}
}

func (state *validationState) hasProtectedWildcard(roleID uuid.UUID) bool {
	for _, permission := range state.snapshot.Permissions {
		if normalizedTarget(permission.Target) != "role" || permission.RoleID == nil ||
			*permission.RoleID != roleID {
			continue
		}
		resource := state.resources[permission.ResourceID]
		if canonicalCatalogKey(resource.Key) != "all" {
			continue
		}
		for _, actionID := range permission.ActionIDs {
			if canonicalCatalogKey(state.actions[actionID].Key) == "all" {
				return true
			}
		}
	}
	return false
}
