package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	organizationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/organization"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
)

func (port *identitySystemResourcePort) createSystemOrganization(
	command identitySystemResourceCommand,
	claim identitySystemResourceCreateClaim,
	values identitySystemResourceValues,
) (publicresource.MutationResult, error) {
	name, _, nameErr := values.text(publicresource.OrganizationNameFieldID, true)
	slug, _, slugErr := values.text(publicresource.OrganizationSlugFieldID, true)
	parentID, _, parentErr := values.nullableUUID(publicresource.OrganizationParentIDFieldID)
	owners, _, ownerErr := values.uuidList(publicresource.OrganizationOwnerIDsFieldID, true)
	if nameErr != nil || slugErr != nil || parentErr != nil || ownerErr != nil ||
		len(owners) == 0 {
		return publicresource.MutationResult{}, publicresource.ErrContract
	}
	profile, err := port.createNameProfile(
		command, identityfacade.ProfileOrganization, values, name,
	)
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	operationID := command.operation(claim.idempotencyKey)
	requestSHA, err := identitySystemResourceCreateDigest(
		claim.resource, claim.values, [32]byte{},
	)
	if err != nil {
		return publicresource.MutationResult{}, publicresource.ErrContract
	}
	evidence := identityMutationEvidence(
		command.evidence, command.permissionID, operationID, requestSHA,
	)
	organizationID := identityChildID(operationID, "organization")
	ownerIDs, membershipIDs := identityOrganizationOwnerCommand(operationID, owners)
	result, err := port.application.identities.organizations.CreateOrganization(
		command.ctx, command.scope, organizationapp.CreateOrganizationCommand{
			ID: identityfacade.IdentityID(organizationID), Name: name, Slug: slug,
			ParentID:           identityfacade.IdentityID(parentID),
			OwnerIDs:           ownerIDs,
			OwnerMembershipIDs: membershipIDs,
			ProfileValues:      profile,
			Evidence:           evidence,
		},
	)
	if err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	if err = port.application.finishIdentityMutation(
		command.ctx, command.scope, evidence, organizationID, "organization",
		command.authorityOperation, result.Replayed,
	); err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	return port.mutationResult(
		command, organizationID, operationID, requestSHA,
		result.Revision, result.Replayed,
	)
}
