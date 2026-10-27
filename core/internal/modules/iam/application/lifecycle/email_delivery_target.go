package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) EmailDeliveryTarget(
	ctx context.Context,
	scope platformtx.TenantScope,
	query EmailDeliveryTargetQuery,
) (DeliveryTarget, error) {
	if !service.Valid() || !tenantSnapshotUsable(ctx, scope) ||
		query.Evidence.Validate() != nil || scope.ActorID() != query.Evidence.ActorID ||
		query.IdentifierID == uuid.Nil {
		return DeliveryTarget{}, ErrInvalid
	}
	return service.validDeliveryTarget(
		service.store.EmailDeliveryTarget(ctx, scope, query),
	)
}
