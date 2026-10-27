package item

import (
	"context"

	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
)

func (storage *store) RequestPurge(ctx context.Context,
	input itemapp.Mutation,
) (itemapp.Snapshot, bool, error) {
	if !validMutation(ctx, storage, input) || len(input.Reason) < 3 ||
		input.StepUpAt.IsZero() {
		return itemapp.Snapshot{}, false, itemapp.ErrInvalid
	}
	var document []byte
	var replay, legalHold bool
	err := storage.queryRow(ctx, requestPurgeSQL, input.OperationID.String(),
		input.AuditID.String(), input.ActorID.String(), input.ItemID.String(),
		input.ExpectedRowVersion, input.IdempotencyKey, input.RequestSHA256[:],
		input.OccurredAt, input.Reason, input.StepUpAt).Scan(&document, &replay, &legalHold)
	if err != nil {
		if storage.notFound(err) {
			return itemapp.Snapshot{}, false, itemapp.ErrConflict
		}
		return itemapp.Snapshot{}, false, itemapp.ErrUnavailable
	}
	if legalHold {
		return itemapp.Snapshot{}, false, itemapp.ErrLegalHold
	}
	result, err := decodeItem(storage.tenantID, document)
	return result, replay, err
}
