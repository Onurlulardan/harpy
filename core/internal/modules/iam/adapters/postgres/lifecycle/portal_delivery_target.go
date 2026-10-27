package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) PortalDeliveryTarget(ctx context.Context,
	scope tenantScope, query lifecycle.PortalDeliveryTargetQuery,
) (lifecycle.DeliveryTarget, error) {
	if !snapshotUsable(ctx, scope) || query.Evidence.Validate() != nil ||
		scope.ActorID() != query.Evidence.ActorID {
		return lifecycle.DeliveryTarget{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return lifecycle.DeliveryTarget{}, err
	}
	e := query.Evidence
	return scanDeliveryTarget(bound.database.QueryRow(ctx, portalDeliveryTargetSQL,
		query.BindingID, e.ActorID, e.PermissionID, e.ExpectedTokenVersion,
		e.ExpectedAuthorizationVersion))
}
