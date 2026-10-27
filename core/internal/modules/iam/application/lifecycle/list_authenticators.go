package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) ListAuthenticators(
	ctx context.Context, scope platformtx.TenantScope, query AuthenticatorListQuery,
) ([]Authenticator, error) {
	if !service.Valid() || !tenantSnapshotUsable(ctx, scope) ||
		query.UserID == uuid.Nil || query.ActorID != scope.ActorID() ||
		query.PermissionID == uuid.Nil || query.ExpectedTokenVersion < 1 ||
		query.ExpectedAuthorizationVersion < 1 || query.Limit < 1 || query.Limit > 100 ||
		(query.AfterUpdatedAt.IsZero() != (query.AfterID == uuid.Nil)) {
		return nil, ErrInvalid
	}
	return service.store.ListAuthenticators(ctx, scope, query)
}
