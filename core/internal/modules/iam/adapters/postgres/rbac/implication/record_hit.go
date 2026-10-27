package implicationpostgres

import (
	"context"
	_ "embed"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

//go:embed sql/tenant/record_hit.sql
var recordHitSQL string

func (storage *store) RecordHit(ctx context.Context, scope tenantScope, command rbacdomain.RecordHitCommand) (rbacdomain.MutationResult, error) {
	if command.Validate() != nil || !scopeUsable(ctx, scope, commandMode) || scope.ActorID() != databaseID(command.Evidence.ActorID) {
		return rbacdomain.MutationResult{}, rbacdomain.ErrInvalidMutation
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return rbacdomain.MutationResult{}, err
	}
	e := command.Evidence
	return scanMutation(bound.database.QueryRow(ctx, recordHitSQL, databaseID(e.OperationID), e.RequestSHA256[:], databaseID(command.ID), command.ExpectedRevision,
		databaseID(e.ActorID), databaseID(e.RequiredPermissionID), e.ExpectedTokenVersion, e.ExpectedAuthorizationVersion))
}
