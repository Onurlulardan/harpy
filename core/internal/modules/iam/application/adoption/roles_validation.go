package adoption

import (
	"strings"

	"github.com/google/uuid"
)

func (state *validationState) validateRoles() {
	names := make(map[string]uuid.UUID)
	defaults := make(map[string]uuid.UUID)
	for _, role := range state.snapshot.Roles {
		key := role.ID.String()
		if role.ID == uuid.Nil || role.TenantID != state.snapshot.TenantID {
			state.add(issue("tenant_or_identity_invalid", "roles", key))
		}
		if _, duplicate := state.roles[role.ID]; duplicate {
			state.add(issue("duplicate_identity", "roles", key))
		}
		state.roles[role.ID] = role
		scope := "tenant"
		if role.OrganizationID != "" {
			organizationID, err := uuid.Parse(role.OrganizationID)
			if err != nil || state.orgs[organizationID].ID == uuid.Nil {
				state.add(issue("dangling_organization", "roles", key))
			} else {
				scope = organizationID.String()
			}
		}
		nameKey := scope + "\x00" + strings.TrimSpace(role.Name)
		if prior, duplicate := names[nameKey]; strings.TrimSpace(role.Name) == "" ||
			(duplicate && prior != role.ID) {
			state.add(issue("role_name_collision", "roles", key))
		}
		names[nameKey] = role.ID
		if role.Default {
			if _, duplicate := defaults[scope]; duplicate {
				state.add(issue("default_role_collision", "roles", scope))
			}
			defaults[scope] = role.ID
		}
	}
}
