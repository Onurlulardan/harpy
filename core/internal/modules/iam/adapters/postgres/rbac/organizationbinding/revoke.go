package organizationbindingpostgres

import (
	"context"
	_ "embed"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

//go:embed sql/tenant/revoke.sql
var revokeSQL string

func (storage *store) Revoke(ctx context.Context, scope tenantScope, command rbacdomain.BindingCommand) (rbacdomain.MutationResult, error) {
	if command.ValidateRevoke() != nil || !scopeUsable(ctx, scope, commandMode) || scope.ActorID() != databaseID(command.Evidence.ActorID) {
		return rbacdomain.MutationResult{}, rbacdomain.ErrInvalidMutation
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return rbacdomain.MutationResult{}, err
	}
	e := command.Evidence
	return scanMutation(bound.database.QueryRow(ctx, revokeSQL, databaseID(e.OperationID), e.RequestSHA256[:], databaseID(command.ID),
		databaseID(command.SubjectID), databaseID(command.PermissionID), command.ExpectedRevision,
		databaseID(e.ActorID), databaseID(e.RequiredPermissionID), e.ExpectedTokenVersion, e.ExpectedAuthorizationVersion))
}
