package assignmentpostgres

import (
	"context"
	_ "embed"

	identityapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/get.sql
var getAssignmentSQL string

func (storage *store) GetAssignment(
	ctx context.Context, scope tenantScope, query identityapp.GetAssignmentQuery,
) (identityapp.Assignment, error) {
	if query.Validate() != nil || !scopeUsable(ctx, scope, snapshotMode) ||
		scope.ActorID() != databaseID(query.Evidence.ActorID) {
		return identityapp.Assignment{}, identityapp.ErrInvalidQuery
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return identityapp.Assignment{}, err
	}
	evidence := query.Evidence
	return scanAssignment(bound.database.QueryRow(
		ctx, getAssignmentSQL, databaseID(query.ID), databaseID(evidence.ActorID), databaseID(evidence.PermissionID),
		evidence.ExpectedTokenVersion, evidence.ExpectedAuthorizationVersion,
	))
}
