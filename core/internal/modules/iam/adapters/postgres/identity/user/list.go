package userpostgres

import (
	"context"
	_ "embed"

	identityapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/list.sql
var listUsersSQL string

func (storage *store) ListUsers(
	ctx context.Context, scope tenantScope, query identityapp.ListUsersQuery,
) ([]identityapp.User, error) {
	if query.Validate() != nil || !scopeUsable(ctx, scope, snapshotMode) ||
		scope.ActorID() != databaseID(query.Evidence.ActorID) {
		return nil, identityapp.ErrInvalidQuery
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return nil, err
	}
	evidence := query.Evidence
	rows, err := bound.database.Query(
		ctx, listUsersSQL, nullableID(query.AfterID), query.Limit,
		databaseID(evidence.ActorID), databaseID(evidence.PermissionID), evidence.ExpectedTokenVersion,
		evidence.ExpectedAuthorizationVersion,
	)
	if err != nil {
		return nil, identityapp.ErrStorage
	}
	defer rows.Close()
	result := make([]identityapp.User, 0, query.Limit)
	for rows.Next() {
		user, scanErr := scanUser(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		user.Profile, scanErr = storage.profiles.Load(ctx, scope, identityapp.ProfileUser, user.ID)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, user)
	}
	if rows.Err() != nil {
		return nil, identityapp.ErrStorage
	}
	return result, nil
}
