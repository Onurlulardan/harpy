package systemresource

import (
	"encoding/json"
	"strings"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
)

func userWireStatus(status identityfacade.UserStatus) string {
	switch status {
	case identityfacade.UserSuspended, identityfacade.UserDisabled:
		return "INACTIVE"
	case identityfacade.UserTombstoned:
		return "DELETED"
	default:
		return "ACTIVE"
	}
}

func roleWireStatus(status string) string {
	return strings.ToUpper(status)
}

func organizationWireStatus(status identityfacade.OrganizationStatus) string {
	if status == identityfacade.OrganizationTombstoned {
		return "DELETED"
	}
	return strings.ToUpper(string(status))
}

func userWireAccountType(modes []string) (string, bool) {
	operator, portal := false, false
	for _, mode := range modes {
		operator = operator || mode == "operator_console"
		portal = portal || mode == "portal"
	}
	switch {
	case operator && portal:
		return "operator_portal", true
	case operator:
		return "operator", true
	case portal:
		return "portal", true
	default:
		return "", false
	}
}

func putSingleIdentityProjection(
	values map[publicresource.ID]json.RawMessage,
	visible map[publicresource.ID]struct{},
	fieldID string,
	identities []identityfacade.IdentityID,
) error {
	if len(identities) != 1 || identities[0] == (identityfacade.IdentityID{}) {
		return nil
	}
	return putIdentityValue(values, visible, fieldID, identities[0])
}
