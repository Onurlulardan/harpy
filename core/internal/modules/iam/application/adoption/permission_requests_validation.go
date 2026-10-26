package adoption

import (
	"strings"

	"github.com/google/uuid"
)

func (state *validationState) validatePermissionRequests() {
	mappings := make(map[uuid.UUID]PermissionRequestMapping)
	groups := make(map[uuid.UUID]uuid.UUID)
	requests := make(map[uuid.UUID]uuid.UUID)
	for _, mapping := range state.mapping.PermissionRequests {
		if _, duplicate := mappings[mapping.SourceRequestID]; duplicate {
			state.add(issue("mapping_duplicate", "permission_requests", mapping.SourceRequestID.String()))
		}
		mappings[mapping.SourceRequestID] = mapping
		state.claimRequestID(groups, mapping.GroupID, mapping.SourceRequestID, "group_id_collision")
		state.claimRequestID(requests, mapping.RequestID, mapping.SourceRequestID, "request_id_collision")
	}
	seen := make(map[uuid.UUID]bool)
	for _, source := range state.snapshot.PermissionRequests {
		key := source.ID.String()
		if source.ID == uuid.Nil || seen[source.ID] {
			state.add(issue("request_identity_invalid", "permission_requests", key))
		}
		seen[source.ID] = true
		if state.users[source.RequesterID].ID == uuid.Nil || state.users[source.TargetID].ID == uuid.Nil ||
			state.resources[source.ResourceID].ID == uuid.Nil || state.actions[source.ActionID].ID == uuid.Nil {
			state.add(issue("request_reference_dangling", "permission_requests", key))
		}
		if !validRequestState(source.State) {
			state.add(issue("request_state_invalid", "permission_requests", key))
		}
		mapping, exists := mappings[source.ID]
		if !exists || mapping.GroupID == uuid.Nil || mapping.RequestID == uuid.Nil {
			state.add(issue("request_mapping_missing", "permission_requests", key))
		}
	}
	for sourceID := range mappings {
		if !seen[sourceID] {
			state.add(issue("mapping_source_missing", "permission_requests", sourceID.String()))
		}
	}
}

func (state *validationState) claimRequestID(
	seen map[uuid.UUID]uuid.UUID, id, sourceID uuid.UUID, code string,
) {
	if id == uuid.Nil || sourceID == uuid.Nil {
		state.add(issue("request_mapping_invalid", "permission_requests", sourceID.String()))
		return
	}
	if prior, duplicate := seen[id]; duplicate && prior != sourceID {
		state.add(issue(code, "permission_requests", id.String()))
	}
	seen[id] = sourceID
}

func validRequestState(value string) bool {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "PENDING", "APPROVED", "REJECTED":
		return true
	default:
		return false
	}
}
