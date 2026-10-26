package archive

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Get(ctx context.Context, scope platformtx.TenantScope,
	request GetRequest,
) (Snapshot, error) {
	if application == nil || !request.ArchiveID.UUIDv7() ||
		scope.OperationID() != OperationGet {
		return Snapshot{}, ErrInvalid
	}
	if _, err := application.authorizer.Authorize(ctx, scope, PermissionRead); err != nil {
		return Snapshot{}, ErrForbidden
	}
	store, identity, err := application.binder.Bind(ctx, scope, SnapshotAccess)
	if err != nil || store == nil || !identity.Valid() {
		return Snapshot{}, ErrUnavailable
	}
	result, err := store.Get(ctx, request.ArchiveID)
	if err != nil {
		return Snapshot{}, mapError(err)
	}
	if result.ID != request.ArchiveID || result.ActorID != identity.ActorID {
		return Snapshot{}, ErrNotFound
	}
	return result, nil
}
