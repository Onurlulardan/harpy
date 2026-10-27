package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) GetSessions(
	ctx context.Context, scope platformtx.TenantScope, query SessionListQuery,
) ([]Session, error) {
	if !service.Valid() || !tenantSnapshotUsable(ctx, scope) ||
		query.UserID == uuid.Nil || query.ActorID != scope.ActorID() ||
		query.PermissionID == uuid.Nil || query.ExpectedTokenVersion < 1 ||
		query.ExpectedAuthorizationVersion < 1 || query.Limit < 1 || query.Limit > 100 ||
		(query.AfterLastSeenAt.IsZero() != (query.AfterID == uuid.Nil)) {
		return nil, ErrInvalid
	}
	return service.store.GetSessions(ctx, scope, query)
}
