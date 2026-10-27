package systemresource

import (
	"encoding/json"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
)

func OrganizationRecord(
	organization identityfacade.Organization,
	catalog publicresource.FieldCatalog,
) (publicresource.Record, error) {
	if organization.Validate() != nil {
		return publicresource.Record{}, publicresource.ErrContract
	}
	visible, err := visibleFieldSet(catalog)
	if err != nil {
		return publicresource.Record{}, err
	}
	values := make(map[publicresource.ID]json.RawMessage, len(visible))
	createdAt, createdErr := timestampText(organization.CreatedAt)
	updatedAt, updatedErr := timestampText(organization.UpdatedAt)
	writes := []error{
		putIdentityValue(values, visible, publicresource.OrganizationIDFieldID,
			organization.ID),
		putValue(values, visible, publicresource.OrganizationNameFieldID,
			organization.Name),
		putValue(values, visible, publicresource.OrganizationSlugFieldID,
			organization.Slug),
		putValue(values, visible, publicresource.OrganizationStatusFieldID,
			organizationWireStatus(organization.Status)),
		putIdentityValue(values, visible, publicresource.OrganizationParentIDFieldID,
			organization.ParentID),
		putValue(values, visible, publicresource.OrganizationOwnerIDsFieldID,
			identityTexts(organization.OwnerIDs)),
		putSingleIdentityProjection(values, visible,
			publicresource.OrganizationOwnerIDFieldID, organization.OwnerIDs),
		putIdentityValue(values, visible, publicresource.OrganizationProfileFieldID,
			organization.Profile.SchemaID()),
		putValue(values, visible, publicresource.OrganizationRevisionFieldID,
			organization.Revision),
		putValue(values, visible, publicresource.OrganizationCreatedAtFieldID, createdAt),
		putValue(values, visible, publicresource.OrganizationUpdatedAtFieldID, updatedAt),
		createdErr, updatedErr,
	}
	for _, writeErr := range writes {
		if writeErr != nil {
			return publicresource.Record{}, publicresource.ErrContract
		}
	}
	if err := addProfileValues(
		values, organization.Profile, identityfacade.ProfileOrganization, visible,
	); err != nil {
		return publicresource.Record{}, err
	}
	fieldValues, err := publicresource.NewFieldValues(values)
	if err != nil {
		return publicresource.Record{}, err
	}
	return publicresource.NewRecord(
		publicresource.Organizations, publicresource.ID(organization.ID),
		organization.Revision, fieldValues,
	)
}
