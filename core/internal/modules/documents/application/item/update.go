package item

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Update(ctx context.Context, scope platformtx.TenantScope,
	request MetadataRequest,
) (Snapshot, error) {
	description, _, tagsJSON, metadataOK := foundation.NormalizeItemMetadata(
		request.Description, request.Tags)
	if !request.ItemID.UUIDv7() || request.RowVersion == 0 ||
		!foundation.ValidIdempotency(request.IdempotencyKey) || !metadataOK {
		return Snapshot{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, Command,
		OperationUpdate, PermissionUpdate)
	if err != nil {
		return Snapshot{}, err
	}
	mutation, err := application.newMutation(identity, request.ItemID,
		[16]byte{}, request.RowVersion, description, tagsJSON, "",
		request.IdempotencyKey, OperationUpdate, application.clock.Now())
	if err != nil {
		return Snapshot{}, err
	}
	return applyMutation(ctx, store.Update, mutation)
}
