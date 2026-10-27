package iam

import (
	"encoding/json"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	"github.com/google/uuid"
)

func (values identitySystemResourceValues) profile() ([]byte, error) {
	profile := make(map[string]json.RawMessage)
	for fieldID, value := range values.values {
		if identitySystemCoreField(values.resource, fieldID) {
			continue
		}
		profile[fieldID.String()] = append(json.RawMessage(nil), value...)
	}
	encoded, err := json.Marshal(profile)
	if err != nil {
		return nil, publicresource.ErrContract
	}
	return encoded, nil
}

func (values identitySystemResourceValues) profileWithDisplayName(
	fieldID identityfacade.IdentityID,
	displayName string,
) ([]byte, error) {
	key := uuid.UUID(fieldID).String()
	parsed, err := publicresource.ParseID(key)
	if err != nil || identitySystemCoreField(values.resource, parsed) ||
		displayName == "" {
		return nil, publicresource.ErrContract
	}
	if raw, found := values.values[parsed]; found {
		var supplied string
		if json.Unmarshal(raw, &supplied) != nil || supplied != displayName {
			return nil, publicresource.ErrContract
		}
	}
	profile, err := values.profile()
	if err != nil {
		return nil, err
	}
	var document map[string]json.RawMessage
	if json.Unmarshal(profile, &document) != nil || document == nil {
		return nil, publicresource.ErrContract
	}
	document[key], err = json.Marshal(displayName)
	if err != nil {
		return nil, publicresource.ErrContract
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		return nil, publicresource.ErrContract
	}
	return encoded, nil
}

func (values identitySystemResourceValues) passwordTerms() []string {
	result := make([]string, 0, len(values.values))
	for fieldID, raw := range values.values {
		if identitySystemCoreField(values.resource, fieldID) {
			continue
		}
		var value string
		if json.Unmarshal(raw, &value) == nil && value != "" {
			result = append(result, value)
		}
	}
	return result
}

func identitySystemCoreField(
	resource publicresource.Resource,
	fieldID publicresource.ID,
) bool {
	_, found := identitySystemCoreFields[resource][fieldID.String()]
	return found
}

var identitySystemCoreFields = map[publicresource.Resource]map[string]struct{}{
	publicresource.Users: identitySystemFieldSet([]string{
		publicresource.UserIDFieldID, publicresource.UserEmailFieldID,
		publicresource.UserPortalIdentifierFieldID, publicresource.UserStatusFieldID,
		publicresource.UserIdentityKindFieldID, publicresource.UserRoleIDsFieldID,
		publicresource.UserOrganizationIDsFieldID, publicresource.UserProfileSchemaFieldID,
		publicresource.UserRevisionFieldID, publicresource.UserCreatedAtFieldID,
		publicresource.UserUpdatedAtFieldID,
	}),
	publicresource.Roles: identitySystemFieldSet([]string{
		publicresource.RoleIDFieldID, publicresource.RoleScopeKindFieldID,
		publicresource.RoleScopeIDFieldID, publicresource.RoleOrganizationIDFieldID,
		publicresource.RoleNameFieldID, publicresource.RoleDescriptionFieldID,
		publicresource.RoleStatusFieldID, publicresource.RoleProtectedFieldID,
		publicresource.RoleDefaultFieldID, publicresource.RoleUserIDsFieldID,
		publicresource.RoleProfileSchemaFieldID, publicresource.RoleRevisionFieldID,
		publicresource.RoleCreatedAtFieldID, publicresource.RoleUpdatedAtFieldID,
	}),
	publicresource.Organizations: identitySystemFieldSet([]string{
		publicresource.OrganizationIDFieldID, publicresource.OrganizationNameFieldID,
		publicresource.OrganizationSlugFieldID, publicresource.OrganizationStatusFieldID,
		publicresource.OrganizationParentIDFieldID, publicresource.OrganizationUserIDsFieldID,
		publicresource.OrganizationRoleIDsFieldID, publicresource.OrganizationProfileFieldID,
		publicresource.OrganizationRevisionFieldID,
		publicresource.OrganizationCreatedAtFieldID,
		publicresource.OrganizationUpdatedAtFieldID,
	}),
}

func identitySystemFieldSet(fields []string) map[string]struct{} {
	result := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		result[field] = struct{}{}
	}
	return result
}
