package item

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Restore(ctx context.Context, scope platformtx.TenantScope,
	request MutationRequest,
) (Snapshot, error) {
	store, identity, err := application.mutationScope(ctx, scope, request,
		OperationRestore, PermissionDelete)
	if err != nil {
		return Snapshot{}, err
	}
	mutation, err := application.newMutation(identity, request.ItemID, [16]byte{},
		request.RowVersion, "", nil, "", request.IdempotencyKey,
		OperationRestore, application.clock.Now())
	if err != nil {
		return Snapshot{}, err
	}
	return applyMutation(ctx, store.Restore, mutation)
}
