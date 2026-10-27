package platformiampostgres

import (
	"context"
	_ "embed"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
)

//go:embed sql/tenant/prepare_owner_transfer.sql
var prepareOwnerTransferSQL string

func (storage *store) CreateOwnerTransfer(
	ctx context.Context, scope tenantScope, command platformdomain.OwnerTransferPrepare,
) (platformdomain.OwnerTransferPrepared, error) {
	if !commandScopeUsable(ctx, scope) || scope.ActorID() != databaseID(command.ActorID) {
		return platformdomain.OwnerTransferPrepared{}, platformdomain.ErrInvalidCommand
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return platformdomain.OwnerTransferPrepared{}, err
	}
	return scanOwnerTransferPrepared(bound.database.QueryRow(ctx, prepareOwnerTransferSQL,
		databaseID(command.OperationID), command.RequestSHA256[:], databaseID(command.ActorID),
		command.ActorExpectedRevision, command.ExpectedTokenVersion,
		command.ExpectedAuthorizationVersion, command.ExpectedPlatformClock,
		command.StepUp.PurposeSHA256[:], command.StepUp.AuthTime, command.StepUp.ACR,
		command.StepUp.AMRPassword, command.StepUp.AMRWebAuthn, command.StepUp.AMRTOTP,
		databaseID(command.TargetUserID), command.TargetExpectedRevision,
		command.TargetExpectedTokenVersion, command.TargetExpectedAuthorizationVersion,
		databaseID(command.SuccessorSubjectID), databaseID(command.SuccessorBindingID),
		databaseID(command.RegistryBindingID), command.ExpectedOwnerPolicyRevision,
		command.ExpectedHeadGeneration, command.ExpectedHeadSHA256[:], command.StateSHA256[:],
		command.ProposalSHA256[:], command.ProposalSignerKID,
		command.ProposalSignatureSHA256[:], command.ProposalExpiresAt))
}
