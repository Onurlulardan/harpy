package assignmentpostgres

import (
	"context"
	_ "embed"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	"github.com/google/uuid"
)

//go:embed sql/tenant/list.sql
var listSQL string

func (storage *store) List(
	ctx context.Context,
	scope tenantScope,
	query rbacdomain.AssignmentListQuery,
) ([]rbacdomain.Assignment, error) {
	if query.Validate() != nil || !scopeUsable(ctx, scope, snapshotMode) ||
		scope.ActorID() != databaseID(query.Evidence.ActorID) {
		return nil, rbacdomain.ErrInvalidQuery
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return nil, err
	}
	var afterID *uuid.UUID
	if query.AfterID.Valid() {
		value := databaseID(query.AfterID)
		afterID = &value
	}
	rows, err := bound.database.Query(ctx, listSQL,
		query.State, afterID, query.Limit,
		databaseID(query.Evidence.ActorID),
		databaseID(query.Evidence.RequiredPermissionID),
	)
	if err != nil {
		return nil, mapError(err, rbacdomain.ErrInvalidQuery)
	}
	defer rows.Close()
	result := make([]rbacdomain.Assignment, 0, query.Limit)
	for rows.Next() {
		item, scanErr := scanAssignment(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, mapError(err, rbacdomain.ErrInvalidQuery)
	}
	return result, nil
}
