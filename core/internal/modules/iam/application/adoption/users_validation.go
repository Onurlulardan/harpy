package adoption

import (
	"strings"

	"github.com/google/uuid"
)

func (state *validationState) validateUsers() {
	mappings := make(map[uuid.UUID]UserMapping)
	for _, mapping := range state.mapping.Identity.Users {
		if _, duplicate := mappings[mapping.UserID]; duplicate {
			state.add(issue("mapping_duplicate", "users", mapping.UserID.String()))
		}
		mappings[mapping.UserID] = mapping
	}
	emails := make(map[string]uuid.UUID)
	portals := make(map[string]uuid.UUID)
	for _, user := range state.snapshot.Users {
		key := user.ID.String()
		if user.ID == uuid.Nil || user.TenantID != state.snapshot.TenantID || user.TokenVersion < 1 {
			state.add(issue("tenant_or_identity_invalid", "users", key))
		}
		if _, duplicate := state.users[user.ID]; duplicate {
			state.add(issue("duplicate_identity", "users", key))
		}
		state.users[user.ID] = user
		mapping, mapped := mappings[user.ID]
		if !mapped {
			state.add(issue("mapping_missing", "users", key))
		}
		state.validateUserEmail(user, mapping, emails)
		state.validateUserPortal(user, mapping, portals)
		state.validateUserRelations(user, mapping)
		state.validateCredential(user)
	}
	for userID := range mappings {
		if state.users[userID].ID == uuid.Nil {
			state.add(issue("mapping_source_missing", "users", userID.String()))
		}
	}
	state.validateMappedIdentityCollisions()
}

func (state *validationState) validateUserEmail(
	user SourceUser, mapping UserMapping, seen map[string]uuid.UUID,
) {
	if strings.TrimSpace(user.Email) == "" {
		if user.AccountType != "portal" {
			state.add(issue("email_missing", "identifiers", user.ID.String()))
		}
		if mapping.EmailIdentifierID != nil {
			state.add(issue("email_mapping_unexpected", "identifiers", user.ID.String()))
		}
		return
	}
	value, err := state.normalizers.email(user.Email)
	if err != nil {
		state.add(issue("email_invalid", "identifiers", user.ID.String()))
		return
	}
	if prior, collision := seen[value]; collision && prior != user.ID {
		state.add(issue("email_collision", "identifiers", value))
	}
	seen[value] = user.ID
	if nilUUID(mapping.EmailIdentifierID) {
		state.add(issue("email_mapping_missing", "identifiers", user.ID.String()))
	}
}

func nilUUID(value *uuid.UUID) bool {
	return value == nil || *value == uuid.Nil
}
