package adoption

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const operationRecordHit = "catalog.compatibility.record_hit"

func (application *Application) RecordHit(ctx context.Context, scope platformtx.TenantScope,
	request RouteRequest,
) error {
	if !routePattern.MatchString(request.RouteKey) || request.RequiredConsumers < 0 ||
		!canonicalTime(request.RemoveAfter) {
		return ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeCommand,
		operationRecordHit, "catalog.read")
	if err != nil {
		return err
	}
	now, err := application.now()
	if err != nil || !request.RemoveAfter.After(now) {
		return ErrInvalid
	}
	return store.RecordHit(ctx, request, now)
}
