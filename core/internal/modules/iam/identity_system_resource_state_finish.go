package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	"github.com/google/uuid"
)

func (port *identitySystemResourcePort) finishSystemResourceState(
	command identitySystemResourceCommand,
	id uuid.UUID,
	operationID uuid.UUID,
	requestSHA [32]byte,
	revision int64,
	replayed bool,
	evidence identityfacade.MutationEvidence,
) (publicresource.MutationResult, error) {
	resourceKind := map[publicresource.Resource]string{
		publicresource.Users: "user", publicresource.Roles: "role",
		publicresource.Organizations: "organization",
	}[command.resource]
	if err := port.application.finishIdentityMutation(
		command.ctx, command.scope, evidence, id, resourceKind,
		command.authorityOperation, replayed,
	); err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	return port.mutationResult(
		command, id, operationID, requestSHA, revision, replayed,
	)
}
