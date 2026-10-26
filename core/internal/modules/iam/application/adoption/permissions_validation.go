package adoption

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type expandedPermission struct {
	sourceID   uuid.UUID
	subjectKey string
	resourceID uuid.UUID
	actionID   uuid.UUID
	wildcard   bool
}

func (state *validationState) validatePermissions() {
	state.validateResourceCatalog()
	state.validateActionCatalog()
	mappings := state.permissionMappings()
	seen := make(map[string]uuid.UUID)
	for _, source := range state.snapshot.Permissions {
		for _, expanded := range state.expandPermission(source) {
			key := fmt.Sprintf("%s:%s:%s", expanded.subjectKey,
				expanded.resourceID, expanded.actionID)
			if prior, duplicate := seen[key]; duplicate && prior != source.ID {
				state.add(issue("permission_collision", "permissions", key))
			}
			seen[key] = source.ID
			state.validatePermissionMapping(expanded, mappings)
		}
	}
	for key, values := range mappings {
		if !values.used {
			state.add(issue("mapping_source_missing", "permissions", key))
		}
	}
	state.validatePermissionRequests()
}

func (state *validationState) expandPermission(source SourcePermission) []expandedPermission {
	if source.ID == uuid.Nil || source.TenantID != state.snapshot.TenantID {
		state.add(issue("tenant_or_identity_invalid", "permissions", source.ID.String()))
	}
	subject, subjectID, valid := state.permissionSubject(source)
	if !valid {
		return nil
	}
	resources, resourceWildcard := state.permissionResources(source)
	actions, actionWildcard := state.permissionActions(source)
	if len(resources) == 0 || len(actions) == 0 {
		return nil
	}
	wildcard := resourceWildcard || actionWildcard
	if wildcard && (subject != "role" || subjectID != state.mapping.Identity.ProtectedRoleID) {
		state.add(issue("wildcard_review_required", "permissions", source.ID.String()))
	}
	result := make([]expandedPermission, 0, len(resources)*len(actions))
	for _, resourceID := range resources {
		for _, actionID := range actions {
			result = append(result, expandedPermission{
				sourceID: source.ID, subjectKey: subject + ":" + subjectID.String(),
				resourceID: resourceID, actionID: actionID, wildcard: wildcard,
			})
		}
	}
	return result
}

func normalizedTarget(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
