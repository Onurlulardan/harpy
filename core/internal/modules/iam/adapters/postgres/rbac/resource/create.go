package resourcepostgres

import (
	"context"
	_ "embed"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

//go:embed sql/tenant/create.sql
var createSQL string

func (storage *store) Create(ctx context.Context, scope tenantScope,
	input rbacdomain.ResourceInput) (rbacdomain.MutationResult, error) {
	if input.Validate() != nil || !scopeUsable(ctx, scope, commandMode) ||
		scope.ActorID() != databaseID(input.Evidence.ActorID) {
		return rbacdomain.MutationResult{}, rbacdomain.ErrInvalidMutation
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return rbacdomain.MutationResult{}, err
	}
	evidence := input.Evidence
	return scanMutation(bound.database.QueryRow(ctx, createSQL,
		databaseID(evidence.OperationID), evidence.RequestSHA256[:], databaseID(input.ID),
		input.Key, input.OwnerDomain, input.Kind, input.DisplayName, input.Description,
		input.Protected, databaseID(evidence.ActorID), databaseID(evidence.RequiredPermissionID),
		evidence.ExpectedTokenVersion, evidence.ExpectedAuthorizationVersion))
}
