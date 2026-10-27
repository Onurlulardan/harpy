package platformiampostgres

import (
	"context"
	_ "embed"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	"github.com/google/uuid"
)

//go:embed sql/tenant/authorize.sql
var authorizeSQL string

func (storage *store) Authorize(
	ctx context.Context, scope tenantScope, query platformdomain.EvaluateQuery,
) (platformdomain.Decision, error) {
	if !queryScopeUsable(ctx, scope) ||
		scope.ActorID() != databaseID(query.ActorID) {
		return platformdomain.Decision{}, platformdomain.ErrInvalidQuery
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return platformdomain.Decision{}, err
	}
	var subject uuid.NullUUID
	var decision platformdomain.Decision
	err = bound.database.QueryRow(ctx, authorizeSQL, databaseID(query.ActorID),
		query.ExpectedTokenVersion, query.ExpectedAuthorizationVersion,
		query.ExpectedPlatformClock, query.Capability).Scan(
		&decision.Allowed, &decision.Capability, &decision.DeploymentMode,
		&subject, &decision.PlatformClock, &decision.AuthorizationVersion,
		&decision.Reason)
	if err != nil {
		return platformdomain.Decision{}, storeError(err)
	}
	if subject.Valid {
		decision.SubjectID = domainID(subject.UUID)
	}
	return decision, nil
}
