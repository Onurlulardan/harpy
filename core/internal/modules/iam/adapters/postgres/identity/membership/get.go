package membershippostgres

import (
	"context"
	_ "embed"

	identityapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/get.sql
var getMembershipSQL string

func (storage *store) GetMembership(
	ctx context.Context, scope tenantScope, query identityapp.GetMembershipQuery,
) (identityapp.Membership, error) {
	if query.Validate() != nil || !scopeUsable(ctx, scope, snapshotMode) ||
		scope.ActorID() != databaseID(query.Evidence.ActorID) {
		return identityapp.Membership{}, identityapp.ErrInvalidQuery
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return identityapp.Membership{}, err
	}
	evidence := query.Evidence
	return scanMembership(bound.database.QueryRow(
		ctx, getMembershipSQL, databaseID(query.ID), databaseID(evidence.ActorID), databaseID(evidence.PermissionID),
		evidence.ExpectedTokenVersion, evidence.ExpectedAuthorizationVersion,
	))
}
