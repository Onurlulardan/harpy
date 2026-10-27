package userpostgres

import (
	"context"
	_ "embed"

	identityapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/update.sql
var updateUserSQL string

func (storage *store) UpdateUser(
	ctx context.Context, scope tenantScope, command identityapp.UpdateUserCommand,
) (identityapp.MutationResult, error) {
	profile := command.Profile()
	if command.Validate() != nil || !scopeUsable(ctx, scope, commandMode) ||
		scope.ActorID() != databaseID(command.Evidence.ActorID) ||
		profile.SchemaID() == (identityapp.IdentityID{}) {
		return identityapp.MutationResult{}, identityapp.ErrInvalidCommand
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return identityapp.MutationResult{}, err
	}
	digest := profile.SHA256()
	return scanMutation(bound.database.QueryRow(
		ctx, updateUserSQL, databaseID(command.Evidence.OperationID),
		command.Evidence.RequestSHA256[:], databaseID(command.ID), command.ExpectedRevision,
		databaseID(profile.SchemaID()), profile.Values(), digest[:],
		databaseID(command.Evidence.ActorID), databaseID(command.Evidence.PermissionID),
		command.Evidence.ExpectedTokenVersion,
		command.Evidence.ExpectedAuthorizationVersion,
	))
}
