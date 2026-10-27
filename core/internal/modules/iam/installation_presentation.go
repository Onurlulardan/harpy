package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	resourceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemresource"
	"github.com/google/uuid"
)

type InstallationPresentationCatalogs struct {
	User, Organization, Role publicresource.FieldCatalog
}

func BuildInstallationPresentationCatalogs(
	operationID [16]byte,
) (InstallationPresentationCatalogs, error) {
	if zeroInstallationID(operationID) {
		return InstallationPresentationCatalogs{}, ErrInstallationIdentityInvalid
	}
	ids := installationIDs(operationID)
	user, err := resourceapp.BuildInstallationUserFieldCatalog(
		uuid.UUID(ids.userSchemaV2), uuid.UUID(ids.userField),
		uuid.UUID(ids.userFirstName), uuid.UUID(ids.userLastName),
		uuid.UUID(ids.userPhone),
	)
	if err != nil {
		return InstallationPresentationCatalogs{}, ErrInstallationIdentityInvalid
	}
	organization, err := resourceapp.BuildInstallationFieldCatalog(
		uuid.UUID(ids.organizationSchema), uuid.UUID(ids.organizationField),
		publicresource.Organizations,
	)
	if err != nil {
		return InstallationPresentationCatalogs{}, ErrInstallationIdentityInvalid
	}
	role, err := resourceapp.BuildInstallationFieldCatalog(
		uuid.UUID(ids.roleSchema), uuid.UUID(ids.roleField), publicresource.Roles,
	)
	if err != nil {
		return InstallationPresentationCatalogs{}, ErrInstallationIdentityInvalid
	}
	return InstallationPresentationCatalogs{
		User: user, Organization: organization, Role: role,
	}, nil
}
