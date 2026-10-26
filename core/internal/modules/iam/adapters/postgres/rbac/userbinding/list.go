package userbindingpostgres

import (
	"context"
	_ "embed"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

//go:embed sql/tenant/list.sql
var listSQL string

func (storage *store) List(ctx context.Context, scope tenantScope, query rbacdomain.BindingListQuery) ([]rbacdomain.Binding, error) {
	if query.Validate() != nil || !scopeUsable(ctx, scope, snapshotMode) || scope.ActorID() != databaseID(query.Evidence.ActorID) {
		return nil, rbacdomain.ErrInvalidQuery
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return nil, err
	}
	rows, err := bound.database.Query(ctx, listSQL, databaseID(query.SubjectID), query.State, query.Limit, databaseID(query.Evidence.ActorID), databaseID(query.Evidence.RequiredPermissionID))
	if err != nil {
		return nil, rbacdomain.ErrStorage
	}
	defer rows.Close()
	result := make([]rbacdomain.Binding, 0, query.Limit)
	for rows.Next() {
		item, scanErr := scanBinding(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, item)
	}
	if rows.Err() != nil {
		return nil, rbacdomain.ErrStorage
	}
	return result, nil
}
