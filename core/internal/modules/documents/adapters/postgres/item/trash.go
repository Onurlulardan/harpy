package item

import (
	"context"

	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
)

func (storage *store) Trash(ctx context.Context,
	input itemapp.Mutation,
) (itemapp.Snapshot, bool, error) {
	if !validMutation(ctx, storage, input) {
		return itemapp.Snapshot{}, false, itemapp.ErrInvalid
	}
	return scanMutation(storage.queryRow(ctx, trashSQL, input.OperationID.String(),
		input.AuditID.String(), input.ActorID.String(), input.ItemID.String(),
		input.ExpectedRowVersion, input.IdempotencyKey, input.RequestSHA256[:],
		input.OccurredAt), storage.tenantID, storage.notFound)
}
