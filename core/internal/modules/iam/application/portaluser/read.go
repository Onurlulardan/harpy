package portaluser

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) Read(
	ctx context.Context, scope platformtx.TenantScope, userID uuid.UUID,
) (View, error) {
	if service == nil || ctx == nil || ctx.Err() != nil || userID == uuid.Nil ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsScope(ctx, scope) ||
		(platformtx.RequireActive(scope, platformtx.SnapshotQuery) != nil &&
			platformtx.RequireActive(scope, platformtx.Command) != nil) {
		return View{}, ErrInvalid
	}
	result, err := service.store.Read(ctx, scope, userID)
	if err != nil {
		return View{}, err
	}
	if !result.Valid() || result.ID != userID {
		return View{}, ErrUnavailable
	}
	return result, nil
}
