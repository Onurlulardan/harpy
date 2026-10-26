package platformiampostgres

import (
	"context"
	_ "embed"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
)

//go:embed sql/tenant/commit_owner_transfer.sql
var commitOwnerTransferSQL string

func (storage *store) CompleteOwnerTransfer(
	ctx context.Context, scope tenantScope, command platformdomain.OwnerTransferCommit,
) (platformdomain.OwnerTransferCommitted, error) {
	if !commandScopeUsable(ctx, scope) || scope.ActorID() != databaseID(command.ActorID) {
		return platformdomain.OwnerTransferCommitted{}, platformdomain.ErrInvalidCommand
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return platformdomain.OwnerTransferCommitted{}, err
	}
	return scanOwnerTransferCommitted(bound.database.QueryRow(ctx, commitOwnerTransferSQL,
		databaseID(command.OperationID), command.RequestSHA256[:], databaseID(command.ActorID),
		command.ExpectedTokenVersion, command.ExpectedAuthorizationVersion,
		command.ExpectedPlatformClock, databaseID(command.OutboxEventID), command.EventSHA256[:]))
}
