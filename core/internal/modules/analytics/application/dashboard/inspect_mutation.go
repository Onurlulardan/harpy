package dashboard

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

// InspectUpdate loads the current dashboard under update authority without mutation.
func (application *Application) InspectUpdate(
	ctx context.Context, scope platformtx.TenantScope, dashboardID uuid.UUID,
) (Snapshot, error) {
	return application.inspectMutation(
		ctx, scope, dashboardID, OperationUpdate, PermissionUpdate,
	)
}

// InspectDelete loads the current dashboard under delete authority without mutation.
func (application *Application) InspectDelete(
	ctx context.Context, scope platformtx.TenantScope, dashboardID uuid.UUID,
) (Snapshot, error) {
	return application.inspectMutation(
		ctx, scope, dashboardID, OperationDelete, PermissionDelete,
	)
}

func (application *Application) inspectMutation(
	ctx context.Context, scope platformtx.TenantScope, dashboardID uuid.UUID,
	operation string, permission Permission,
) (Snapshot, error) {
	if !validUUIDv7(dashboardID) {
		return Snapshot{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeCommand, operation, permission)
	if err != nil {
		return Snapshot{}, err
	}
	result, err := store.Get(ctx, dashboardID)
	if err != nil {
		return Snapshot{}, err
	}
	if result.DashboardID != dashboardID || validateSnapshot(result) != nil {
		return Snapshot{}, ErrUnavailable
	}
	return result, nil
}
