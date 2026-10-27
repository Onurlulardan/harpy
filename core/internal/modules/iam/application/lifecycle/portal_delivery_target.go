package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) PortalDeliveryTarget(
	ctx context.Context,
	scope platformtx.TenantScope,
	query PortalDeliveryTargetQuery,
) (DeliveryTarget, error) {
	if !service.Valid() || !tenantSnapshotUsable(ctx, scope) ||
		query.Evidence.Validate() != nil || scope.ActorID() != query.Evidence.ActorID ||
		query.BindingID == uuid.Nil {
		return DeliveryTarget{}, ErrInvalid
	}
	return service.validDeliveryTarget(
		service.store.PortalDeliveryTarget(ctx, scope, query),
	)
}
