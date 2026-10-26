package usage

import (
	"context"
	"errors"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const currentUsageReadOperation = "billing.usage.read"

func (queries *Queries) Get(
	ctx context.Context,
	scope platformtx.TenantScope,
) (UsageSnapshot, error) {
	if queries == nil || queries.current == nil || !usageQueryScope(ctx, scope) {
		return UsageSnapshot{}, ErrQueryScope
	}
	result, err := queries.current.Get(ctx)
	if err != nil {
		return UsageSnapshot{}, normalizeUsageQueryError(err)
	}
	if result.TenantID() != scope.TenantID() {
		return UsageSnapshot{}, ErrQueryInvariant
	}
	if _, err = NewUsageSnapshot(result.TenantID(), result.Meters()); err != nil {
		return UsageSnapshot{}, ErrQueryInvariant
	}
	return result, nil
}

func usageQueryScope(ctx context.Context, scope platformtx.TenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) == nil &&
		scope.OperationID() == currentUsageReadOperation
}

func normalizeUsageQueryError(err error) error {
	for _, expected := range []error{
		ErrQueryDependency, ErrQueryInvariant, ErrQueryStorage,
	} {
		if errors.Is(err, expected) {
			return expected
		}
	}
	return ErrQueryStorage
}
