package platformiampostgres

import (
	"context"
	_ "embed"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
)

//go:embed sql/tenant/prepare_owner_transfer_compensation.sql
var prepareOwnerTransferCompensationSQL string

func (storage *store) ReserveOwnerTransferCompensation(
	ctx context.Context, scope tenantScope, command platformdomain.OwnerTransferCompensation,
) (platformdomain.OwnerTransferPrepared, error) {
	if !commandScopeUsable(ctx, scope) || scope.ActorID() != databaseID(command.ActorID) {
		return platformdomain.OwnerTransferPrepared{}, platformdomain.ErrInvalidCommand
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return platformdomain.OwnerTransferPrepared{}, err
	}
	return scanOwnerTransferPrepared(bound.database.QueryRow(ctx, prepareOwnerTransferCompensationSQL,
		databaseID(command.OperationID), command.RequestSHA256[:], databaseID(command.ActorID),
		command.ExpectedTokenVersion, command.ExpectedAuthorizationVersion,
		command.ExpectedPlatformClock, command.StepUp.PurposeSHA256[:], command.StepUp.AuthTime,
		command.StepUp.ACR, command.StepUp.AMRPassword, command.StepUp.AMRWebAuthn,
		command.StepUp.AMRTOTP, databaseID(command.OriginalOperationID),
		command.OriginalAckSHA256[:], command.ExpectedOriginalGeneration,
		command.ProposalSHA256[:], command.ProposalSignerKID,
		command.ProposalSignatureSHA256[:], command.ProposalExpiresAt))
}
