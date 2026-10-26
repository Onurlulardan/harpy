package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) ReadRecoveryCodes(
	ctx context.Context, scope platformtx.TenantScope, query RecoveryListQuery,
) ([]RecoveryCode, error) {
	if !service.Valid() || !tenantSnapshotUsable(ctx, scope) ||
		query.UserID == uuid.Nil || query.ActorID != scope.ActorID() ||
		query.PermissionID == uuid.Nil || query.ExpectedTokenVersion < 1 ||
		query.ExpectedAuthorizationVersion < 1 {
		return nil, ErrInvalid
	}
	return service.store.ReadRecoveryCodes(ctx, scope, query)
}
