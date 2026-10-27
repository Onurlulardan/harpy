package iam

import (
	"strings"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
	"github.com/google/uuid"
)

func rbacMutationFromRequest(rbacID uuid.UUID, key string, revision int64) RBACMutation {
	return RBACMutation{
		ID: rbacID, ExpectedRevision: revision, IdempotencyKey: key,
	}
}

func rbacTarget(
	target string,
	userID *string,
	roleID *string,
	organizationID *string,
) (rbacdomain.AssignmentTarget, uuid.UUID, bool) {
	kind := rbacdomain.AssignmentTarget(target)
	values := map[rbacdomain.AssignmentTarget]*string{
		rbacdomain.AssignmentUser:         userID,
		rbacdomain.AssignmentRole:         roleID,
		rbacdomain.AssignmentOrganization: organizationID,
	}
	selected, present := values[kind]
	if !kind.Valid() || !present || selected == nil {
		return "", uuid.Nil, false
	}
	for candidate, value := range values {
		if candidate != kind && value != nil {
			return "", uuid.Nil, false
		}
	}
	id, valid := decodeUUID(*selected)
	return kind, id, valid
}

func rbacTargetFromView(view RBACPermissionView) (rbacdomain.AssignmentTarget, uuid.UUID, bool) {
	return rbacTarget(view.Target, view.UserID, view.RoleID, view.OrganizationID)
}

func rbacActionIDsFromView(view RBACPermissionView) ([]uuid.UUID, bool) {
	values := make([]string, len(view.Actions))
	for index, action := range view.Actions {
		values[index] = action.ID
	}
	return decodeUUIDList(values)
}

func rbacCompatibilitySlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	separator := false
	for _, raw := range value {
		mapped := rbacASCII(raw)
		if mapped != 0 && (mapped >= 'a' && mapped <= 'z' ||
			mapped >= '0' && mapped <= '9') {
			if separator && builder.Len() > 0 {
				builder.WriteByte('-')
			}
			builder.WriteRune(mapped)
			separator = false
		} else if builder.Len() > 0 {
			separator = true
		}
	}
	result := strings.Trim(builder.String(), "-")
	if result != "" && result[0] >= '0' && result[0] <= '9' {
		result = "legacy-" + result
	}
	return result
}

func rbacASCII(value rune) rune {
	switch value {
	case 'ç':
		return 'c'
	case 'ğ':
		return 'g'
	case 'ı', 'i':
		return 'i'
	case 'ö':
		return 'o'
	case 'ş':
		return 's'
	case 'ü':
		return 'u'
	default:
		return value
	}
}
