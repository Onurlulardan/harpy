package membershippostgres

import (
	"context"
	_ "embed"

	identityapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/add.sql
var addMembershipSQL string

func (storage *store) AddMembership(
	ctx context.Context, scope tenantScope, command identityapp.AddMembershipCommand,
) (identityapp.MutationResult, error) {
	if command.Validate() != nil || !scopeUsable(ctx, scope, commandMode) ||
		scope.ActorID() != databaseID(command.Evidence.ActorID) {
		return identityapp.MutationResult{}, identityapp.ErrInvalidCommand
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return identityapp.MutationResult{}, err
	}
	evidence := command.Evidence
	return scanMutation(bound.database.QueryRow(
		ctx, addMembershipSQL, databaseID(evidence.OperationID), evidence.RequestSHA256[:],
		databaseID(command.ID), databaseID(command.OrganizationID), databaseID(command.UserID), string(command.Kind),
		databaseID(evidence.ActorID), databaseID(evidence.PermissionID), evidence.ExpectedTokenVersion,
		evidence.ExpectedAuthorizationVersion,
	))
}
