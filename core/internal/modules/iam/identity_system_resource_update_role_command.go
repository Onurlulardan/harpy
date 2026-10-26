package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	roleapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/role"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
)

func (port *identitySystemResourcePort) updateSystemRoleCommand(
	command identitySystemResourceCommand,
	update identitySystemResourceUpdate,
	name string,
	description string,
	defaultValue bool,
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
	result, err := port.application.identities.roles.UpdateRole(
		command.ctx, command.scope, roleapp.UpdateRoleCommand{
			ID:               identityfacade.IdentityID(update.id),
			ExpectedRevision: update.expectedRevision,
			Name:             name, Description: description, Default: defaultValue,
			ProfilePatch: profile, Evidence: evidence,
		},
	)
	if err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	if err = port.application.finishIdentityMutation(
		command.ctx, command.scope, evidence, update.id, "role",
		command.authorityOperation, result.Replayed,
	); err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	return port.mutationResult(
		command, update.id, operationID, requestSHA,
		result.Revision, result.Replayed,
	)
}
