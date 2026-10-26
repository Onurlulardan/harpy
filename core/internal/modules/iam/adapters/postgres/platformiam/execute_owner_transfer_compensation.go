package platformiampostgres

import (
	"context"
	_ "embed"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
)

//go:embed sql/tenant/apply_owner_transfer_compensation.sql
var applyOwnerTransferCompensationSQL string

func (storage *store) ExecuteOwnerTransferCompensation(
	ctx context.Context, scope tenantScope, command platformdomain.OwnerTransferAcknowledge,
) (platformdomain.OwnerTransferCompensated, error) {
	if !commandScopeUsable(ctx, scope) || scope.ActorID() != databaseID(command.ActorID) {
		return platformdomain.OwnerTransferCompensated{}, platformdomain.ErrInvalidCommand
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return platformdomain.OwnerTransferCompensated{}, err
	}
	ack := command.Acknowledgement
	return scanOwnerTransferCompensated(bound.database.QueryRow(ctx,
		applyOwnerTransferCompensationSQL, databaseID(command.OperationID),
		command.RequestSHA256[:], databaseID(command.ActorID), command.ExpectedTokenVersion,
		command.ExpectedAuthorizationVersion, command.ExpectedPlatformClock,
		command.StepUp.PurposeSHA256[:], command.StepUp.AuthTime, command.StepUp.ACR,
		command.StepUp.AMRPassword, command.StepUp.AMRWebAuthn, command.StepUp.AMRTOTP,
		databaseID(ack.RegistryBindingID), databaseID(ack.ScopeID), ack.RegistrySequence,
		ack.Generation, ack.ProposalSHA256[:], ack.PreviousHeadSHA256[:], ack.StateSHA256[:],
		ack.AckSHA256[:], ack.IssuedAt, ack.RegistryKID))
}
