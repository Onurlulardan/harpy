package systemresource

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	formpresentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
)

func fieldMode(
	resource publicresource.Resource,
	fieldID string,
) formpresentation.FieldMode {
	if editableCoreField(resource, fieldID) {
		return formpresentation.FieldEditable
	}
	if coreField(resource, fieldID) {
		return formpresentation.FieldReadOnly
	}
	return formpresentation.FieldEditable
}

func editableCoreField(resource publicresource.Resource, fieldID string) bool {
	switch resource {
	case publicresource.Users:
		return fieldID == publicresource.UserEmailFieldID ||
			fieldID == publicresource.UserRoleIDsFieldID ||
			fieldID == publicresource.UserOrganizationIDsFieldID
	case publicresource.Roles:
		return fieldID == publicresource.RoleScopeKindFieldID ||
			fieldID == publicresource.RoleScopeIDFieldID ||
			fieldID == publicresource.RoleOrganizationIDFieldID ||
			fieldID == publicresource.RoleNameFieldID ||
			fieldID == publicresource.RoleDescriptionFieldID ||
			fieldID == publicresource.RoleDefaultFieldID
	case publicresource.Organizations:
		return fieldID == publicresource.OrganizationNameFieldID ||
			fieldID == publicresource.OrganizationSlugFieldID ||
			fieldID == publicresource.OrganizationParentIDFieldID ||
			fieldID == publicresource.OrganizationOwnerIDsFieldID
	default:
		return false
	}
}

func coreField(resource publicresource.Resource, fieldID string) bool {
	prefix := map[publicresource.Resource]byte{
		publicresource.Users: 'a', publicresource.Roles: 'b',
		publicresource.Organizations: 'c',
	}[resource]
	return len(fieldID) == 36 && fieldID[0] == prefix &&
		fieldID[1:8] == "1000000"
}
