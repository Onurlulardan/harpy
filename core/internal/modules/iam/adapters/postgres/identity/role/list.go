package rolepostgres

import (
	"context"
	_ "embed"

	identityapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/list.sql
var listRolesSQL string

func (storage *store) ListRoles(
	ctx context.Context, scope tenantScope, query identityapp.ListRolesQuery,
) ([]identityapp.Role, error) {
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
		ctx, listRolesSQL, nullableID(query.AfterID), query.Limit,
		databaseID(evidence.ActorID), databaseID(evidence.PermissionID), evidence.ExpectedTokenVersion,
		evidence.ExpectedAuthorizationVersion,
	)
	if err != nil {
		return nil, identityapp.ErrStorage
	}
	defer rows.Close()
	result := make([]identityapp.Role, 0, query.Limit)
	for rows.Next() {
		role, scanErr := scanRole(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		role.Profile, scanErr = storage.profiles.Load(ctx, scope, identityapp.ProfileRole, role.ID)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, role)
	}
	if rows.Err() != nil {
		return nil, identityapp.ErrStorage
	}
	return result, nil
}
