package publication

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) Get(
	ctx context.Context, scope platformtx.TenantScope, releaseID uuid.UUID,
) (Snapshot, error) {
	if !validUUIDv7(releaseID) {
		return Snapshot{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeSnapshot, OperationGet, PermissionRead)
	if err != nil {
		return Snapshot{}, err
	}
	result, err := store.Get(ctx, releaseID)
	if err != nil {
		return Snapshot{}, err
	}
	if result.ReleaseID != releaseID || validateSnapshot(result) != nil {
		return Snapshot{}, ErrUnavailable
	}
	return result, nil
}
