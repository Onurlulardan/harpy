package organizationpostgres

import (
	"context"
	_ "embed"

	identityapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/list.sql
var listOrganizationsSQL string

func (storage *store) ListOrganizations(
	ctx context.Context, scope tenantScope, query identityapp.ListOrganizationsQuery,
) ([]identityapp.Organization, error) {
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
		ctx, listOrganizationsSQL, nullableID(query.AfterID), query.Limit,
		databaseID(evidence.ActorID), databaseID(evidence.PermissionID), evidence.ExpectedTokenVersion,
		evidence.ExpectedAuthorizationVersion,
	)
	if err != nil {
		return nil, identityapp.ErrStorage
	}
	defer rows.Close()
	result := make([]identityapp.Organization, 0, query.Limit)
	for rows.Next() {
		organization, scanErr := scanOrganization(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		organization.Profile, scanErr = storage.profiles.Load(
			ctx, scope, identityapp.ProfileOrganization, organization.ID,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, organization)
	}
	if rows.Err() != nil {
		return nil, identityapp.ErrStorage
	}
	return result, nil
}
