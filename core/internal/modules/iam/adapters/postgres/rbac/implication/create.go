package implicationpostgres

import (
	"context"
	_ "embed"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

//go:embed sql/tenant/create.sql
var createSQL string

func (storage *store) Create(ctx context.Context, scope tenantScope, input rbacdomain.ImplicationInput) (rbacdomain.MutationResult, error) {
	if input.Validate() != nil || !scopeUsable(ctx, scope, commandMode) || scope.ActorID() != databaseID(input.Evidence.ActorID) {
		return rbacdomain.MutationResult{}, rbacdomain.ErrInvalidMutation
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return rbacdomain.MutationResult{}, err
	}
	e := input.Evidence
	return scanMutation(bound.database.QueryRow(ctx, createSQL, databaseID(e.OperationID), e.RequestSHA256[:], databaseID(input.ID),
		databaseID(input.SourcePermissionID), databaseID(input.ImpliedPermissionID), input.RetirementContractKey,
		databaseID(e.ActorID), databaseID(e.RequiredPermissionID), e.ExpectedTokenVersion, e.ExpectedAuthorizationVersion))
}
