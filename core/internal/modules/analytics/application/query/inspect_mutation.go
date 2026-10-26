package query

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

// InspectUpdate loads the current query under update authority without mutating it.
func (application *Application) InspectUpdate(
	ctx context.Context, scope platformtx.TenantScope, queryID uuid.UUID,
) (Snapshot, error) {
	return application.inspectMutation(ctx, scope, queryID, OperationUpdate, PermissionUpdate)
}

// InspectDelete loads the current query under delete authority without mutating it.
func (application *Application) InspectDelete(
	ctx context.Context, scope platformtx.TenantScope, queryID uuid.UUID,
) (Snapshot, error) {
	return application.inspectMutation(ctx, scope, queryID, OperationDelete, PermissionDelete)
}

func (application *Application) inspectMutation(
	ctx context.Context, scope platformtx.TenantScope, queryID uuid.UUID,
	operation string, permission Permission,
) (Snapshot, error) {
	if !validUUIDv7(queryID) {
		return Snapshot{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeCommand, operation, permission)
	if err != nil {
		return Snapshot{}, err
	}
	result, err := store.Get(ctx, queryID)
	if err != nil {
		return Snapshot{}, err
	}
	if result.QueryID != queryID || validateSnapshot(result) != nil {
		return Snapshot{}, ErrUnavailable
	}
	return result, nil
}
