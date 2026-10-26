package item

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Move(ctx context.Context, scope platformtx.TenantScope,
	request MoveRequest,
) (Snapshot, error) {
	if !request.ItemID.UUIDv7() || !request.FolderID.UUIDv7() || request.RowVersion == 0 ||
		!foundation.ValidIdempotency(request.IdempotencyKey) {
		return Snapshot{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, Command,
		OperationMove, PermissionMove)
	if err != nil {
		return Snapshot{}, err
	}
	mutation, err := application.newMutation(identity, request.ItemID, request.FolderID,
		request.RowVersion, "", nil, "", request.IdempotencyKey,
		OperationMove, application.clock.Now())
	if err != nil {
		return Snapshot{}, err
	}
	return applyMutation(ctx, store.Move, mutation)
}
