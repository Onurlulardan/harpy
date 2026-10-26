package subscriptionapp

import (
	"context"
	"errors"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const currentSubscriptionReadOperation = "billing.subscription.read"

func (queries *Queries) GetCurrent(
	ctx context.Context,
	scope platformtx.TenantScope,
) (CurrentSubscription, error) {
	if queries == nil || queries.current == nil || !currentQueryScope(ctx, scope) {
		return CurrentSubscription{}, ErrQueryScope
	}
	result, err := queries.current.GetCurrent(ctx)
	if err != nil {
		return CurrentSubscription{}, normalizeCurrentQueryError(err)
	}
	if result.TenantID() != scope.TenantID() || !result.valid() {
		return CurrentSubscription{}, ErrCurrentSubscriptionInvariant
	}
	return result, nil
}

func currentQueryScope(ctx context.Context, scope platformtx.TenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) == nil &&
		scope.OperationID() == currentSubscriptionReadOperation
}

func normalizeCurrentQueryError(err error) error {
	for _, expected := range []error{
		ErrCurrentSubscriptionNotFound, ErrCurrentSubscriptionInvariant,
		ErrSubscriptionQueryDependency, ErrSubscriptionQueryStorage,
	} {
		if errors.Is(err, expected) {
			return expected
		}
	}
	return ErrSubscriptionQueryStorage
}
