package assignmentpostgres

import (
	"context"
	_ "embed"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

//go:embed sql/tenant/retire.sql
var retireSQL string

func (storage *store) Retire(
	ctx context.Context,
	scope tenantScope,
	command rbacdomain.AssignmentMutation,
) (rbacdomain.MutationResult, error) {
	if command.ValidateRetire() != nil || !scopeUsable(ctx, scope, commandMode) ||
		scope.ActorID() != databaseID(command.Evidence.ActorID) {
		return rbacdomain.MutationResult{}, rbacdomain.ErrInvalidMutation
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return rbacdomain.MutationResult{}, err
	}
	return scanMutation(bound.database.QueryRow(ctx, retireSQL,
		databaseID(command.Evidence.OperationID), command.Evidence.RequestSHA256[:],
		databaseID(command.ID), command.ExpectedRevision,
		databaseID(command.Evidence.ActorID),
		databaseID(command.Evidence.RequiredPermissionID),
		command.Evidence.ExpectedTokenVersion,
		command.Evidence.ExpectedAuthorizationVersion,
		command.StepUp.PurposeSHA256[:], command.StepUp.AuthTime,
		command.StepUp.AMRPassword, command.StepUp.AMRWebAuthn,
		command.StepUp.AMRTOTP,
	))
}
