package permissionpostgres

import (
	"context"
	_ "embed"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

//go:embed sql/tenant/get.sql
var getSQL string

func (storage *store) Get(ctx context.Context, scope tenantScope, query rbacdomain.GetQuery) (rbacdomain.Permission, error) {
	if query.Validate() != nil || !scopeUsable(ctx, scope, snapshotMode) || scope.ActorID() != databaseID(query.Evidence.ActorID) {
		return rbacdomain.Permission{}, rbacdomain.ErrInvalidQuery
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return rbacdomain.Permission{}, err
	}
	return scanPermission(bound.database.QueryRow(ctx, getSQL, databaseID(query.ID), databaseID(query.Evidence.ActorID), databaseID(query.Evidence.RequiredPermissionID)))
}
