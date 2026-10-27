package userpostgres

import (
	"context"
	_ "embed"

	identityapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/get.sql
var getUserSQL string

func (storage *store) GetUser(
	ctx context.Context, scope tenantScope, query identityapp.GetUserQuery,
) (identityapp.User, error) {
	if query.Validate() != nil || !scopeUsable(ctx, scope, snapshotMode) ||
		scope.ActorID() != databaseID(query.Evidence.ActorID) {
		return identityapp.User{}, identityapp.ErrInvalidQuery
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return identityapp.User{}, err
	}
	evidence := query.Evidence
	result, err := scanUser(bound.database.QueryRow(
		ctx, getUserSQL, databaseID(query.ID), databaseID(evidence.ActorID), databaseID(evidence.PermissionID),
		evidence.ExpectedTokenVersion, evidence.ExpectedAuthorizationVersion,
	))
	if err != nil {
		return identityapp.User{}, err
	}
	result.Profile, err = storage.profiles.Load(ctx, scope, identityapp.ProfileUser, result.ID)
	if err != nil {
		return identityapp.User{}, err
	}
	return result, nil
}
