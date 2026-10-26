package dashboardexport

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) Get(
	ctx context.Context, scope platformtx.TenantScope, dashboardID, exportID uuid.UUID,
) (Snapshot, error) {
	if !validUUIDv7(dashboardID) || !validUUIDv7(exportID) {
		return Snapshot{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeSnapshot, OperationGet)
	if err != nil {
		return Snapshot{}, err
	}
	now := application.clock.Now()
	result, err := store.Get(ctx, dashboardID, exportID, now)
	if err != nil {
		return Snapshot{}, err
	}
	if validateSnapshot(result, dashboardID, exportID) != nil {
		return Snapshot{}, ErrUnavailable
	}
	return result, nil
}
