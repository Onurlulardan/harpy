package portaluser

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) VerifyAccess(
	ctx context.Context,
	scope platformtx.TenantScope,
	principalID, bindingID uuid.UUID,
	tokenVersion, authorizationVersion int64,
) (AccessBinding, error) {
	if service == nil || ctx == nil || ctx.Err() != nil || principalID == uuid.Nil ||
		bindingID == uuid.Nil || tokenVersion < 1 || authorizationVersion < 1 ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) != nil ||
		scope.ActorID() != principalID {
		return AccessBinding{}, ErrInvalid
	}
	result, err := service.store.VerifyAccess(
		ctx, scope, principalID, bindingID, tokenVersion, authorizationVersion,
	)
	if err != nil {
		return AccessBinding{}, err
	}
	if !result.Valid() || result.PrincipalID != principalID ||
		result.BindingID != bindingID || result.TenantID != scope.TenantID() {
		return AccessBinding{}, ErrUnavailable
	}
	return result, nil
}
