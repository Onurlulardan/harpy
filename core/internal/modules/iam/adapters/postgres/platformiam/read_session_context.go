package platformiampostgres

import (
	"context"
	_ "embed"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	"github.com/google/uuid"
)

//go:embed sql/tenant/read_session_context.sql
var readSessionContextSQL string

func (storage *store) ReadSessionContext(
	ctx context.Context, scope tenantScope, query platformdomain.SessionContextQuery,
) (platformdomain.SessionContext, error) {
	if !queryScopeUsable(ctx, scope) ||
		scope.ActorID() != databaseID(query.ActorID) {
		return platformdomain.SessionContext{}, platformdomain.ErrInvalidQuery
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return platformdomain.SessionContext{}, err
	}
	var result platformdomain.SessionContext
	var subject uuid.NullUUID
	err = bound.database.QueryRow(ctx, readSessionContextSQL, databaseID(query.ActorID),
		query.ExpectedTokenVersion, query.ExpectedAuthorizationVersion).Scan(
		&result.DeploymentMode, &result.TenantPurpose, &subject,
		&result.PlatformClock, &result.AuthorizationVersion)
	if err != nil {
		return platformdomain.SessionContext{}, storeError(err)
	}
	if subject.Valid {
		result.SubjectID = domainID(subject.UUID)
	}
	return result, nil
}
