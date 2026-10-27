package userpostgres

import (
	"context"
	_ "embed"

	identityapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/delete.sql
var deleteUserSQL string

//go:embed sql/platform/directory_tombstone.sql
var directoryTombstoneSQL string

func (storage *store) DeleteUser(
	ctx context.Context, scope tenantScope, command identityapp.UserStateCommand,
) (identityapp.MutationResult, error) {
	if command.Validate() != nil || !scopeUsable(ctx, scope, commandMode) ||
		scope.ActorID() != databaseID(command.Evidence.ActorID) {
		return identityapp.MutationResult{}, identityapp.ErrInvalidCommand
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return identityapp.MutationResult{}, err
	}
	authority, err := iamAuthority(scope)
	if err != nil {
		return identityapp.MutationResult{}, err
	}
	evidence := command.Evidence
	result, err := scanMutation(bound.database.QueryRow(
		ctx, deleteUserSQL, databaseID(evidence.OperationID), evidence.RequestSHA256[:],
		databaseID(command.ID), command.ExpectedRevision,
		databaseID(evidence.ActorID), databaseID(evidence.PermissionID), evidence.ExpectedTokenVersion,
		evidence.ExpectedAuthorizationVersion,
	))
	if err != nil {
		return identityapp.MutationResult{}, err
	}
	var directoryRevision int64
	if err := bound.database.QueryRow(
		ctx, directoryTombstoneSQL, scope.TenantID(), authority.GroupGeneration(),
		authority.Generation(), databaseID(command.ID), result.Revision,
	).Scan(&directoryRevision); err != nil || directoryRevision != result.Revision {
		return identityapp.MutationResult{}, identityapp.ErrStorage
	}
	return result, nil
}
