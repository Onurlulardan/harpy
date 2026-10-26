package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	organizationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/organization"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
)

func (port *identitySystemResourcePort) updateSystemOrganization(
	command identitySystemResourceCommand,
	update identitySystemResourceUpdate,
	values identitySystemResourceValues,
) (publicresource.MutationResult, error) {
	current, err := port.application.identities.organizations.GetOrganization(
		command.ctx, command.scope, organizationapp.GetOrganizationQuery{
			ID:       identityfacade.IdentityID(update.id),
			Evidence: identityQueryEvidence(command.evidence, command.readPermissionID),
		},
	)
	if err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	if current.Revision != update.expectedRevision {
		return publicresource.MutationResult{}, publicresource.ErrConflict
	}
	name, slug, parentID := current.Name, current.Slug, current.ParentID
	if value, found, fieldErr := values.text(publicresource.OrganizationNameFieldID, false); fieldErr != nil {
		return publicresource.MutationResult{}, fieldErr
	} else if found {
		name = value
	}
	if value, found, fieldErr := values.text(publicresource.OrganizationSlugFieldID, false); fieldErr != nil {
		return publicresource.MutationResult{}, fieldErr
	} else if found {
		slug = value
	}
	if value, found, fieldErr := values.nullableUUID(publicresource.OrganizationParentIDFieldID); fieldErr != nil {
		return publicresource.MutationResult{}, fieldErr
	} else if found {
		parentID = identityfacade.IdentityID(value)
	}
	owners, replaceOwners, err := values.uuidList(
		publicresource.OrganizationOwnerIDsFieldID, false,
	)
	if err != nil || replaceOwners && len(owners) == 0 {
		return publicresource.MutationResult{}, publicresource.ErrContract
	}
	profile, err := port.updateNameProfile(
		command, identityfacade.ProfileOrganization,
		current.Profile.SchemaID(), values, name,
	)
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	return port.updateSystemOrganizationCommand(
		command, update, name, slug, parentID, owners, replaceOwners, profile,
	)
}
