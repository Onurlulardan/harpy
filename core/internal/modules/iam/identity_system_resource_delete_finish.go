package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	"github.com/google/uuid"
)

func (port *identitySystemResourcePort) finishSystemResourceDelete(
	command identitySystemResourceCommand,
	id uuid.UUID,
	operationID uuid.UUID,
	requestSHA [32]byte,
	revision int64,
	replayed bool,
	evidence identityfacade.MutationEvidence,
) (publicresource.DeleteResult, error) {
	resourceKind := map[publicresource.Resource]string{
		publicresource.Users: "user", publicresource.Roles: "role",
		publicresource.Organizations: "organization",
	}[command.resource]
	if err := port.application.finishIdentityMutation(
		command.ctx, command.scope, evidence, id, resourceKind,
		command.authorityOperation, replayed,
	); err != nil {
		return publicresource.DeleteResult{}, identitySystemResourceError(err)
	}
	receipt, err := publicresource.NewMutationReceipt(
		publicresource.ID(operationID), requestSHA, revision, replayed,
	)
	if err != nil {
		return publicresource.DeleteResult{}, publicresource.ErrUnavailable
	}
	result, err := publicresource.NewDeleteResult(
		command.resource, command.catalog, receipt,
	)
	if err != nil {
		return publicresource.DeleteResult{}, publicresource.ErrUnavailable
	}
	return result, nil
}
