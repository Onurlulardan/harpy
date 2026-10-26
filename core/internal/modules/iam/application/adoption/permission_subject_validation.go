package adoption

import "github.com/google/uuid"

func (state *validationState) permissionSubject(
	source SourcePermission,
) (string, uuid.UUID, bool) {
	target := normalizedTarget(source.Target)
	values := map[string]*uuid.UUID{
		"user": source.UserID, "role": source.RoleID,
		"organization": source.OrganizationID,
	}
	nonNil := 0
	for _, value := range values {
		if value != nil && *value != uuid.Nil {
			nonNil++
		}
	}
	selected, known := values[target]
	if !known || selected == nil || *selected == uuid.Nil || nonNil != 1 {
		state.add(issue("permission_subject_invalid", "permissions", source.ID.String()))
		return "", uuid.Nil, false
	}
	if !state.subjectExists(target, *selected) {
		state.add(issue("permission_subject_dangling", "permissions", source.ID.String()))
		return "", uuid.Nil, false
	}
	return target, *selected, true
}

func (state *validationState) subjectExists(kind string, id uuid.UUID) bool {
	switch kind {
	case "user":
		return state.users[id].ID != uuid.Nil
	case "role":
		return state.roles[id].ID != uuid.Nil
	case "organization":
		return state.orgs[id].ID != uuid.Nil
	default:
		return false
	}
}
