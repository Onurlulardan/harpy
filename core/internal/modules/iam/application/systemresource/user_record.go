package systemresource

import (
	"encoding/json"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
)

func UserRecord(
	user identityfacade.User,
	catalog publicresource.FieldCatalog,
) (publicresource.Record, error) {
	if user.Validate() != nil {
		return publicresource.Record{}, publicresource.ErrContract
	}
	visible, err := visibleFieldSet(catalog)
	if err != nil {
		return publicresource.Record{}, err
	}
	values := make(map[publicresource.ID]json.RawMessage, len(visible))
	createdAt, createdErr := timestampText(user.CreatedAt)
	updatedAt, updatedErr := timestampText(user.UpdatedAt)
	accountType, hasAccountType := userWireAccountType(user.AccessModes)
	writes := []error{
		putIdentityValue(values, visible, publicresource.UserIDFieldID, user.ID),
		putValue(values, visible, publicresource.UserEmailFieldID, user.EmailDisplay),
		putValue(values, visible, publicresource.UserStatusFieldID, userWireStatus(user.Status)),
		putValue(values, visible, publicresource.UserIdentityKindFieldID, "human"),
		putValue(values, visible, publicresource.UserRoleIDsFieldID, identityTexts(user.RoleIDs)),
		putValue(values, visible, publicresource.UserOrganizationIDsFieldID,
			identityTexts(user.OrganizationIDs)),
		putSingleIdentityProjection(values, visible,
			publicresource.UserOrganizationIDFieldID, user.OrganizationIDs),
		putValue(values, visible, publicresource.UserEmailVerifiedFieldID,
			user.EmailVerified),
		putIdentityValue(values, visible, publicresource.UserProfileSchemaFieldID,
			user.Profile.SchemaID()),
		putValue(values, visible, publicresource.UserRevisionFieldID, user.Revision),
		putValue(values, visible, publicresource.UserCreatedAtFieldID, createdAt),
		putValue(values, visible, publicresource.UserUpdatedAtFieldID, updatedAt),
		createdErr, updatedErr,
	}
	if hasAccountType {
		writes = append(writes, putValue(
			values, visible, publicresource.UserAccountTypeFieldID, accountType,
		))
	}
	for _, writeErr := range writes {
		if writeErr != nil {
			return publicresource.Record{}, publicresource.ErrContract
		}
	}
	if err := addProfileValues(
		values, user.Profile, identityfacade.ProfileUser, visible,
	); err != nil {
		return publicresource.Record{}, err
	}
	fieldValues, err := publicresource.NewFieldValues(values)
	if err != nil {
		return publicresource.Record{}, err
	}
	return publicresource.NewRecord(
		publicresource.Users, publicresource.ID(user.ID), user.Revision, fieldValues,
	)
}
