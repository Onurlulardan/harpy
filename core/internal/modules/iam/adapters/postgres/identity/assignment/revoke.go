package assignmentpostgres

import (
	"context"
	_ "embed"

	identityapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/revoke.sql
var revokeAssignmentSQL string

func (storage *store) RevokeAssignment(
	ctx context.Context, scope tenantScope, command identityapp.RevokeAssignmentCommand,
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
		ctx, revokeAssignmentSQL, databaseID(evidence.OperationID), evidence.RequestSHA256[:],
		databaseID(command.ID), command.ExpectedRevision, databaseID(evidence.ActorID), databaseID(evidence.PermissionID),
		evidence.ExpectedTokenVersion, evidence.ExpectedAuthorizationVersion,
	))
}
