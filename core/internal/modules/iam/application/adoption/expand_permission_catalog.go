package adoption

import (
	"sort"

	"github.com/google/uuid"
)

func (state *validationState) permissionResources(
	source SourcePermission,
) ([]uuid.UUID, bool) {
	resource, exists := state.resources[source.ResourceID]
	if !exists {
		state.add(issue("permission_resource_dangling", "permissions", source.ID.String()))
		return nil, false
	}
	if canonicalCatalogKey(resource.Key) != "all" {
		return []uuid.UUID{source.ResourceID}, false
	}
	ids := make([]uuid.UUID, 0, len(state.resources))
	for id, value := range state.resources {
		if canonicalCatalogKey(value.Key) != "all" {
			ids = append(ids, id)
		}
	}
	sortUUIDs(ids)
	if len(ids) == 0 {
		state.add(issue("wildcard_expansion_empty", "permissions", source.ID.String()))
	}
	return ids, true
}

func (state *validationState) permissionActions(
	source SourcePermission,
) ([]uuid.UUID, bool) {
	seen := make(map[uuid.UUID]bool)
	wildcard := false
	for _, id := range source.ActionIDs {
		action, exists := state.actions[id]
		if !exists {
			state.add(issue("permission_action_dangling", "permissions", source.ID.String()))
			continue
		}
		if seen[id] {
			state.add(issue("permission_action_duplicate", "permissions", source.ID.String()))
		}
		seen[id] = true
		wildcard = wildcard || canonicalCatalogKey(action.Key) == "all"
	}
	if wildcard {
		ids := make([]uuid.UUID, 0, len(state.actions))
		for id, value := range state.actions {
			if canonicalCatalogKey(value.Key) != "all" {
				ids = append(ids, id)
			}
		}
		sortUUIDs(ids)
		if len(ids) == 0 {
			state.add(issue("wildcard_expansion_empty", "permissions", source.ID.String()))
		}
		return ids, true
	}
	ids := make([]uuid.UUID, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sortUUIDs(ids)
	if len(ids) == 0 {
		state.add(issue("permission_actions_empty", "permissions", source.ID.String()))
	}
	return ids, false
}

func sortUUIDs(values []uuid.UUID) {
	sort.Slice(values, func(left, right int) bool {
		return values[left].String() < values[right].String()
	})
}
