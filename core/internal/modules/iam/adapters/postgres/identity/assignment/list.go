package assignmentpostgres

import (
	"context"
	_ "embed"

	identityapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/list.sql
var listAssignmentsSQL string

func (storage *store) ListAssignments(
	ctx context.Context, scope tenantScope, query identityapp.ListAssignmentsQuery,
) ([]identityapp.Assignment, error) {
	if query.Validate() != nil || !assignmentListScopeUsable(ctx, scope) ||
		scope.ActorID() != databaseID(query.Evidence.ActorID) {
		return nil, identityapp.ErrInvalidQuery
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return nil, err
	}
	evidence := query.Evidence
	rows, err := bound.database.Query(
		ctx, listAssignmentsSQL, nullableID(query.AfterID), query.Limit,
		databaseID(evidence.ActorID), databaseID(evidence.PermissionID), evidence.ExpectedTokenVersion,
		evidence.ExpectedAuthorizationVersion,
	)
	if err != nil {
		return nil, identityapp.ErrStorage
	}
	defer rows.Close()
	result := make([]identityapp.Assignment, 0, query.Limit)
	for rows.Next() {
		assignment, scanErr := scanAssignment(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, assignment)
	}
	if rows.Err() != nil {
		return nil, identityapp.ErrStorage
	}
	return result, nil
}

func assignmentListScopeUsable(ctx context.Context, scope tenantScope) bool {
	return scopeUsable(ctx, scope, snapshotMode) ||
		scope.OperationID() == "iam.identity.user.delete" &&
			scopeUsable(ctx, scope, commandMode)
}
