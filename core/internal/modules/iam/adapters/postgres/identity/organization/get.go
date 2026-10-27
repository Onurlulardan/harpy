package organizationpostgres

import (
	"context"
	_ "embed"

	identityapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/get.sql
var getOrganizationSQL string

func (storage *store) GetOrganization(
	ctx context.Context, scope tenantScope, query identityapp.GetOrganizationQuery,
) (identityapp.Organization, error) {
	if query.Validate() != nil || !scopeUsable(ctx, scope, snapshotMode) ||
		scope.ActorID() != databaseID(query.Evidence.ActorID) {
		return identityapp.Organization{}, identityapp.ErrInvalidQuery
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return identityapp.Organization{}, err
	}
	evidence := query.Evidence
	result, err := scanOrganization(bound.database.QueryRow(
		ctx, getOrganizationSQL, databaseID(query.ID), databaseID(evidence.ActorID), databaseID(evidence.PermissionID),
		evidence.ExpectedTokenVersion, evidence.ExpectedAuthorizationVersion,
	))
	if err != nil {
		return identityapp.Organization{}, err
	}
	result.Profile, err = storage.profiles.Load(ctx, scope, identityapp.ProfileOrganization, result.ID)
	if err != nil {
		return identityapp.Organization{}, err
	}
	return result, nil
}
