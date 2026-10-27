package adoption

import (
	"fmt"

	"github.com/google/uuid"
)

type permissionMappingState struct {
	value PermissionMapping
	used  bool
}

func permissionMappingKey(sourceID, resourceID, actionID uuid.UUID) string {
	return fmt.Sprintf("%s:%s:%s", sourceID, resourceID, actionID)
}

func (state *validationState) permissionMappings() map[string]*permissionMappingState {
	result := make(map[string]*permissionMappingState)
	permissionIDs := make(map[uuid.UUID]string)
	bindingIDs := make(map[uuid.UUID]string)
	for _, mapping := range state.mapping.Permissions {
		key := permissionMappingKey(mapping.SourcePermissionID, mapping.ResourceID, mapping.ActionID)
		if mapping.SourcePermissionID == uuid.Nil || mapping.ResourceID == uuid.Nil ||
			mapping.ActionID == uuid.Nil || mapping.PermissionID == uuid.Nil ||
			mapping.BindingID == uuid.Nil {
			state.add(issue("permission_mapping_invalid", "permissions", key))
		}
		if _, duplicate := result[key]; duplicate {
			state.add(issue("mapping_duplicate", "permissions", key))
		}
		state.claimPermissionID(permissionIDs, mapping.PermissionID, key, "permission_id_collision")
		state.claimPermissionID(bindingIDs, mapping.BindingID, key, "binding_id_collision")
		result[key] = &permissionMappingState{value: mapping}
	}
	return result
}

func (state *validationState) claimPermissionID(
	seen map[uuid.UUID]string, id uuid.UUID, sourceKey, code string,
) {
	if id == uuid.Nil {
		return
	}
	if prior, duplicate := seen[id]; duplicate && prior != sourceKey {
		state.add(issue(code, "permissions", id.String()))
	}
	seen[id] = sourceKey
}

func (state *validationState) validatePermissionMapping(
	expanded expandedPermission, mappings map[string]*permissionMappingState,
) {
	resourceID := state.targetResourceID(expanded.resourceID)
	actionID := state.targetActionID(expanded.actionID)
	if resourceID == uuid.Nil || actionID == uuid.Nil {
		return
	}
	key := permissionMappingKey(expanded.sourceID, resourceID, actionID)
	mapping, exists := mappings[key]
	if !exists {
		state.add(issue("permission_mapping_missing", "permissions", key))
		return
	}
	mapping.used = true
}

func (state *validationState) targetResourceID(sourceID uuid.UUID) uuid.UUID {
	for _, mapping := range state.mapping.Resources {
		if mapping.SourceID == sourceID {
			return mapping.TargetID
		}
	}
	return uuid.Nil
}

func (state *validationState) targetActionID(sourceID uuid.UUID) uuid.UUID {
	for _, mapping := range state.mapping.Actions {
		if mapping.SourceID == sourceID {
			return mapping.TargetID
		}
	}
	return uuid.Nil
}
