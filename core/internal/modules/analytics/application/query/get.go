package query

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) Get(
	ctx context.Context, scope platformtx.TenantScope, queryID uuid.UUID,
) (Snapshot, error) {
	if !validUUIDv7(queryID) {
		return Snapshot{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeSnapshot,
		OperationGet, PermissionRead)
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
