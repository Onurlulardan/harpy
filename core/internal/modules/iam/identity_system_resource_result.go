package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

func (port *identitySystemResourcePort) mutationResult(
	command identitySystemResourceCommand,
	identityID uuid.UUID,
	operationID uuid.UUID,
	requestSHA [32]byte,
	revision int64,
	replayed bool,
) (publicresource.MutationResult, error) {
	record, err := port.readRecord(
		command.ctx, command.scope, command.evidence, command.readPermissionID,
		command.resource, identityID, command.catalog,
	)
	if err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	read, err := publicresource.NewReadResult(record, command.catalog)
	if err != nil {
		return publicresource.MutationResult{}, publicresource.ErrUnavailable
	}
	receipt, err := publicresource.NewMutationReceipt(
		publicresource.ID(operationID), requestSHA, revision, replayed,
	)
	if err != nil {
		return publicresource.MutationResult{}, publicresource.ErrUnavailable
	}
	result, err := publicresource.NewMutationResult(read, receipt)
	if err != nil {
		return publicresource.MutationResult{}, publicresource.ErrUnavailable
	}
	return result, nil
}
