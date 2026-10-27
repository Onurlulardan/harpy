package platformiampostgres

import (
	"context"
	_ "embed"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
)

//go:embed sql/tenant/acknowledge_owner_transfer.sql
var acknowledgeOwnerTransferSQL string

func (storage *store) RecordOwnerTransferAck(
	ctx context.Context, scope tenantScope, command platformdomain.OwnerTransferAcknowledge,
) (platformdomain.OwnerTransferPrepared, error) {
	if !commandScopeUsable(ctx, scope) || scope.ActorID() != databaseID(command.ActorID) {
		return platformdomain.OwnerTransferPrepared{}, platformdomain.ErrInvalidCommand
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return platformdomain.OwnerTransferPrepared{}, err
	}
	ack := command.Acknowledgement
	return scanOwnerTransferPrepared(bound.database.QueryRow(ctx, acknowledgeOwnerTransferSQL,
		databaseID(command.OperationID), command.RequestSHA256[:], databaseID(command.ActorID),
		command.ExpectedTokenVersion, command.ExpectedAuthorizationVersion,
		command.ExpectedPlatformClock, command.StepUp.PurposeSHA256[:], command.StepUp.AuthTime,
		command.StepUp.ACR, command.StepUp.AMRPassword, command.StepUp.AMRWebAuthn,
		command.StepUp.AMRTOTP, databaseID(ack.RegistryBindingID), databaseID(ack.ScopeID),
		ack.RegistrySequence, ack.Generation, ack.ProposalSHA256[:], ack.PreviousHeadSHA256[:],
		ack.StateSHA256[:], ack.AckSHA256[:], ack.IssuedAt, ack.RegistryKID))
}
