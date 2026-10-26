package assignmentpostgres

import (
	"context"
	_ "embed"

	identityapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/update.sql
var updateAssignmentSQL string

func (storage *store) UpdateAssignment(
	ctx context.Context, scope tenantScope, command identityapp.UpdateAssignmentCommand,
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
		ctx, updateAssignmentSQL, databaseID(evidence.OperationID), evidence.RequestSHA256[:],
		databaseID(command.ID), command.ExpectedRevision, string(command.State),
		databaseID(evidence.ActorID), databaseID(evidence.PermissionID), evidence.ExpectedTokenVersion,
		evidence.ExpectedAuthorizationVersion,
	))
}
