package iam

import (
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

func validIdentityUserView(view IdentityUserView) bool {
	return view.ID != uuid.Nil && validIdentityText(view.Email, 254) &&
		validIdentityText(view.FirstName, 200) && validIdentityOptionalText(view.LastName, 200) &&
		validIdentityOptionalText(view.Phone, 32) &&
		(view.AccountType == "" || view.AccountType == "operator" ||
			view.AccountType == "portal" || view.AccountType == "operator_portal") &&
		(view.Status == "ACTIVE" || view.Status == "INACTIVE" || view.Status == "DELETED") &&
		validIdentityIDs(view.RoleIDs) && validIdentityIDs(view.OrganizationIDs) &&
		validIdentityUserOrganizationAlias(view.OrganizationIDs, view.OrganizationID) &&
		validIdentityViewCommon(view.Revision, view.CreatedAt.IsZero(),
			view.UpdatedAt.Before(view.CreatedAt), view.CustomAttributes)
}

func validIdentityUserOrganizationAlias(
	organizationIDs []uuid.UUID,
	organizationID *uuid.UUID,
) bool {
	if len(organizationIDs) != 1 {
		return organizationID == nil
	}
	return organizationID != nil && *organizationID == organizationIDs[0]
}

func validIdentityRoleView(view IdentityRoleView) bool {
	validScope := view.ScopeKind == "TENANT" && view.ScopeID != uuid.Nil &&
		view.OrganizationID == uuid.Nil || view.ScopeKind == "ORGANIZATION" &&
		view.ScopeID != uuid.Nil && view.OrganizationID == view.ScopeID
	return view.ID != uuid.Nil && validScope && validIdentityText(view.Name, 200) &&
		len(view.Description) <= 2000 && utf8.ValidString(view.Description) &&
		(view.Status == "ACTIVE" || view.Status == "RETIRED") &&
		validIdentityViewCommon(view.Revision, view.CreatedAt.IsZero(),
			view.UpdatedAt.Before(view.CreatedAt), view.CustomAttributes)
}

func validIdentityOrganizationView(view IdentityOrganizationView) bool {
	return view.ID != uuid.Nil && view.ParentID != view.ID &&
		validIdentityText(view.Name, 200) && validIdentityText(view.Slug, 100) &&
		(view.Status == "ACTIVE" || view.Status == "SUSPENDED" || view.Status == "DELETED") &&
		validIdentityIDs(view.OwnerIDs) &&
		validIdentityUserOrganizationAlias(view.OwnerIDs, view.OwnerID) &&
		validIdentityViewCommon(view.Revision, view.CreatedAt.IsZero(),
			view.UpdatedAt.Before(view.CreatedAt), view.CustomAttributes)
}

func validIdentityViewCommon(
	revision int64,
	createdZero bool,
	updatedBefore bool,
	custom json.RawMessage,
) bool {
	if revision < 1 || createdZero || updatedBefore {
		return false
	}
	if len(custom) == 0 {
		return true
	}
	var values map[string]json.RawMessage
	if json.Unmarshal(custom, &values) != nil || len(values) > 100 {
		return false
	}
	for key := range values {
		if !validIdentityProfileKey(key) {
			return false
		}
	}
	return true
}

func validIdentityIDs(values []uuid.UUID) bool {
	if values == nil || len(values) > 100 {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			return false
		}
		if _, duplicate := seen[value]; duplicate {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}

func validIdentityText(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && utf8.ValidString(value) &&
		strings.TrimSpace(value) == value
}

func validIdentityOptionalText(value string, maximum int) bool {
	return value == "" || validIdentityText(value, maximum)
}

func validIdentityProfileKey(value string) bool {
	if len(value) < 1 || len(value) > 64 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for _, character := range value[1:] {
		if character != '_' && (character < 'a' || character > 'z') &&
			(character < '0' || character > '9') {
			return false
		}
	}
	return true
}
