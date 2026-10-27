package rolepostgres

import (
	"context"
	_ "embed"

	identityapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/get.sql
var getRoleSQL string

func (storage *store) GetRole(
	ctx context.Context, scope tenantScope, query identityapp.GetRoleQuery,
) (identityapp.Role, error) {
	if query.Validate() != nil || !scopeUsable(ctx, scope, snapshotMode) ||
		scope.ActorID() != databaseID(query.Evidence.ActorID) {
		return identityapp.Role{}, identityapp.ErrInvalidQuery
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return identityapp.Role{}, err
	}
	evidence := query.Evidence
	result, err := scanRole(bound.database.QueryRow(
		ctx, getRoleSQL, databaseID(query.ID), databaseID(evidence.ActorID), databaseID(evidence.PermissionID),
		evidence.ExpectedTokenVersion, evidence.ExpectedAuthorizationVersion,
	))
	if err != nil {
		return identityapp.Role{}, err
	}
	result.Profile, err = storage.profiles.Load(ctx, scope, identityapp.ProfileRole, result.ID)
	if err != nil {
		return identityapp.Role{}, err
	}
	return result, nil
}
