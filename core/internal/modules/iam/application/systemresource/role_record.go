package systemresource

import (
	"encoding/json"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
)

func RoleRecord(
	role identityfacade.Role,
	catalog publicresource.FieldCatalog,
) (publicresource.Record, error) {
	if role.Validate() != nil {
		return publicresource.Record{}, publicresource.ErrContract
	}
	visible, err := visibleFieldSet(catalog)
	if err != nil {
		return publicresource.Record{}, err
	}
	values := make(map[publicresource.ID]json.RawMessage, len(visible))
	createdAt, createdErr := timestampText(role.CreatedAt)
	updatedAt, updatedErr := timestampText(role.UpdatedAt)
	writes := []error{
		putIdentityValue(values, visible, publicresource.RoleIDFieldID, role.ID),
		putValue(values, visible, publicresource.RoleScopeKindFieldID, string(role.ScopeKind)),
		putIdentityValue(values, visible, publicresource.RoleScopeIDFieldID, role.ScopeID),
		putIdentityValue(values, visible, publicresource.RoleOrganizationIDFieldID,
			role.OrganizationID),
		putValue(values, visible, publicresource.RoleNameFieldID, role.Name),
		putValue(values, visible, publicresource.RoleDescriptionFieldID, role.Description),
		putValue(values, visible, publicresource.RoleStatusFieldID,
			roleWireStatus(string(role.Status))),
		putValue(values, visible, publicresource.RoleProtectedFieldID, role.Protected),
		putValue(values, visible, publicresource.RoleDefaultFieldID, role.Default),
		putIdentityValue(values, visible, publicresource.RoleProfileSchemaFieldID,
			role.Profile.SchemaID()),
		putValue(values, visible, publicresource.RoleRevisionFieldID, role.Revision),
		putValue(values, visible, publicresource.RoleCreatedAtFieldID, createdAt),
		putValue(values, visible, publicresource.RoleUpdatedAtFieldID, updatedAt),
		createdErr, updatedErr,
	}
	for _, writeErr := range writes {
		if writeErr != nil {
			return publicresource.Record{}, publicresource.ErrContract
		}
	}
	if err := addProfileValues(
		values, role.Profile, identityfacade.ProfileRole, visible,
	); err != nil {
		return publicresource.Record{}, err
	}
	fieldValues, err := publicresource.NewFieldValues(values)
	if err != nil {
		return publicresource.Record{}, err
	}
	return publicresource.NewRecord(
		publicresource.Roles, publicresource.ID(role.ID), role.Revision, fieldValues,
	)
}
