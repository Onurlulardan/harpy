package item

import (
	"context"

	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
)

func (storage *store) Move(ctx context.Context,
	input itemapp.Mutation,
) (itemapp.Snapshot, bool, error) {
	if !validMutation(ctx, storage, input) || !input.FolderID.UUIDv7() {
		return itemapp.Snapshot{}, false, itemapp.ErrInvalid
	}
	return scanMutation(storage.queryRow(ctx, moveSQL, input.OperationID.String(),
		input.AuditID.String(), input.ActorID.String(), input.ItemID.String(),
		input.ExpectedRowVersion, input.IdempotencyKey, input.RequestSHA256[:],
		input.OccurredAt, input.FolderID.String()), storage.tenantID, storage.notFound)
}
