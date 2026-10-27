package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	organizationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/organization"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	"github.com/google/uuid"
)

func (port *identitySystemResourcePort) updateSystemOrganizationCommand(
	command identitySystemResourceCommand,
	update identitySystemResourceUpdate,
	name string,
	slug string,
	parentID identityfacade.IdentityID,
	owners []uuid.UUID,
	replaceOwners bool,
	profile []byte,
) (publicresource.MutationResult, error) {
	operationID := command.operation(update.idempotencyKey)
	requestSHA, err := identitySystemResourceUpdateDigest(
		command.resource, update.id, update.expectedRevision, update.values,
	)
	if err != nil {
		return publicresource.MutationResult{}, publicresource.ErrContract
	}
	evidence := identityMutationEvidence(
		command.evidence, command.permissionID, operationID, requestSHA,
	)
	ownerIDs, membershipIDs := identityOrganizationOwnerCommand(operationID, owners)
	result, err := port.application.identities.organizations.UpdateOrganization(
		command.ctx, command.scope, organizationapp.UpdateOrganizationCommand{
			ID:               identityfacade.IdentityID(update.id),
			ExpectedRevision: update.expectedRevision,
			Name:             name, Slug: slug, ParentID: parentID,
			ReplaceOwners:      replaceOwners,
			OwnerIDs:           ownerIDs,
			OwnerMembershipIDs: membershipIDs,
			ProfilePatch:       profile,
			Evidence:           evidence,
		},
	)
	if err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	if err = port.application.finishIdentityMutation(
		command.ctx, command.scope, evidence, update.id, "organization",
		command.authorityOperation, result.Replayed,
	); err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	return port.mutationResult(
		command, update.id, operationID, requestSHA,
		result.Revision, result.Replayed,
	)
}
