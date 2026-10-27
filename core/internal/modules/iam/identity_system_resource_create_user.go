package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

func (port *identitySystemResourcePort) createSystemUser(
	command identitySystemResourceCommand,
	claim identitySystemResourceCreateClaim,
	values identitySystemResourceValues,
) (publicresource.MutationResult, error) {
	if !claim.credential.directory.Valid() || !claim.credential.password.Valid() ||
		claim.credential.passwordSHA256 == [32]byte{} {
		return publicresource.MutationResult{}, publicresource.ErrContract
	}
	organizations, _, err := values.uuidList(
		publicresource.UserOrganizationIDsFieldID, true,
	)
	if err != nil || len(organizations) == 0 {
		return publicresource.MutationResult{}, publicresource.ErrContract
	}
	roles, found, err := values.uuidList(publicresource.UserRoleIDsFieldID, false)
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	if !found {
		roles = []uuid.UUID{}
	}
	profile, err := values.profile()
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	grants, err := port.application.userRoleGrantsForOrganizationsInScope(
		command.ctx, command.scope, command.evidence, command.permissionID,
		organizations, true, roles,
	)
	if err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	operationID := command.operation(claim.idempotencyKey)
	requestSHA, err := identitySystemResourceCreateDigest(
		claim.resource, claim.values, claim.credential.passwordSHA256,
	)
	if err != nil {
		return publicresource.MutationResult{}, publicresource.ErrContract
	}
	ids := identityNewUserIDs(command.evidence, claim.idempotencyKey)
	replayed, err := port.application.createUserGraphPrepared(
		command.ctx, command.scope, command.evidence, organizations, profile,
		grants, claim.credential, ids, command.permissionID, requestSHA,
	)
	if err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	evidence := identityMutationEvidence(
		command.evidence, command.permissionID, operationID, requestSHA,
	)
	if err = port.application.finishIdentityMutation(
		command.ctx, command.scope, evidence, ids.user, "user",
		command.authorityOperation, replayed,
	); err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	return port.mutationResult(
		command, ids.user, operationID, requestSHA, 2, replayed,
	)
}
