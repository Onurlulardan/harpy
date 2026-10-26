package permissionpostgres

import (
	"context"
	_ "embed"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

//go:embed sql/tenant/retire.sql
var retireSQL string

func (storage *store) Retire(ctx context.Context, scope tenantScope, command rbacdomain.StateCommand) (rbacdomain.MutationResult, error) {
	if command.Validate() != nil || !scopeUsable(ctx, scope, commandMode) || scope.ActorID() != databaseID(command.Evidence.ActorID) {
		return rbacdomain.MutationResult{}, rbacdomain.ErrInvalidMutation
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return rbacdomain.MutationResult{}, err
	}
	e := command.Evidence
	return scanMutation(bound.database.QueryRow(ctx, retireSQL, databaseID(e.OperationID), e.RequestSHA256[:], databaseID(command.ID),
		command.ExpectedRevision, databaseID(e.ActorID), databaseID(e.RequiredPermissionID), e.ExpectedTokenVersion, e.ExpectedAuthorizationVersion))
}
