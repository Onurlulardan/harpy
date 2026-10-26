package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	"github.com/google/uuid"
)

func identitySystemRoleScopeUnchanged(
	values identitySystemResourceValues,
	current identityfacade.Role,
) error {
	if value, found, err := values.text(publicresource.RoleScopeKindFieldID, false); err != nil {
		return err
	} else if found && value != string(current.ScopeKind) {
		return publicresource.ErrForbidden
	}
	if value, found, err := values.nullableUUID(publicresource.RoleScopeIDFieldID); err != nil {
		return err
	} else if found && value != uuid.UUID(current.ScopeID) {
		return publicresource.ErrForbidden
	}
	if value, found, err := values.nullableUUID(
		publicresource.RoleOrganizationIDFieldID,
	); err != nil {
		return err
	} else if found && value != uuid.UUID(current.OrganizationID) {
		return publicresource.ErrForbidden
	}
	return nil
}
