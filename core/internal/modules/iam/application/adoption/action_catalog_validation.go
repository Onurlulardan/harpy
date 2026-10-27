package adoption

import "github.com/google/uuid"

func (state *validationState) validateActionCatalog() {
	mappings := make(map[uuid.UUID]ActionMapping)
	for _, mapping := range state.mapping.Actions {
		if _, duplicate := mappings[mapping.SourceID]; duplicate {
			state.add(issue("mapping_duplicate", "actions", mapping.SourceID.String()))
		}
		mappings[mapping.SourceID] = mapping
	}
	seen := make(map[string]uuid.UUID)
	for _, source := range state.snapshot.Actions {
		key := canonicalCatalogKey(source.Key)
		if source.ID == uuid.Nil || source.TenantID != state.snapshot.TenantID {
			state.add(issue("tenant_or_identity_invalid", "actions", source.ID.String()))
		}
		if _, duplicate := state.actions[source.ID]; duplicate {
			state.add(issue("duplicate_identity", "actions", source.ID.String()))
		}
		state.actions[source.ID] = source
		if prior, duplicate := seen[key]; key == "" || duplicate && prior != source.ID {
			state.add(issue("action_key_collision", "actions", key))
		}
		seen[key] = source.ID
		mapping := mappings[source.ID]
		if key == "all" {
			if mapping.SourceID != uuid.Nil {
				state.add(issue("wildcard_mapping_unexpected", "actions", source.ID.String()))
			}
		} else if mapping.SourceID == uuid.Nil || mapping.TargetID == uuid.Nil ||
			mapping.ActionKey != key || !catalogKeyPattern.MatchString(mapping.ActionKey) {
			state.add(issue("action_mapping_invalid", "actions", source.ID.String()))
		}
	}
	for sourceID := range mappings {
		if state.actions[sourceID].ID == uuid.Nil {
			state.add(issue("mapping_source_missing", "actions", sourceID.String()))
		}
	}
}
