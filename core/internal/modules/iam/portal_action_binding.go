package iam

import (
	"slices"

	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
)

func samePortalActionBinding(
	current portaluser.AccessBinding, access PortalAccessDecision,
) bool {
	return current.Valid() && access.Valid() &&
		current.PrincipalID == access.PrincipalID() &&
		current.TenantID == access.TenantID() &&
		current.BindingID == access.BindingID() &&
		current.FormID == access.FormID() &&
		current.FormVersionID == access.FormVersionID() &&
		current.SubjectRecordID == access.SubjectRecordID() &&
		current.IdentifierFieldID == access.IdentifierFieldID() &&
		current.SubjectType == access.SubjectType() &&
		slices.Equal(current.RoleNames, access.RoleNames()) &&
		current.TokenVersion == access.TokenVersion() &&
		current.AuthorizationVersion == access.AuthorizationVersion() &&
		current.AuthorizationClock >= access.AuthorizationClock() &&
		current.BindingRevision == access.BindingRevision()
}
