package formscleanup

import (
	"context"

	formscleanupapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/formscleanup"
)

func (storage *store) AcknowledgeEnqueue(ctx context.Context,
	enqueued formscleanupapp.Enqueued,
) error {
	if !storage.usable(ctx) || !enqueued.OperationID.UUIDv7() ||
		!enqueued.ItemID.UUIDv7() || !enqueued.JobID.UUIDv7() ||
		enqueued.RunAt.IsZero() || enqueued.EnqueuedAt.IsZero() {
		return formscleanupapp.ErrInvalid
	}
	var operationID string
	err := storage.queryRow(ctx, markEnqueuedSQL, enqueued.OperationID.String(),
		enqueued.ItemID.String(), enqueued.JobID.String(), enqueued.RunAt,
		enqueued.EnqueuedAt).Scan(&operationID)
	if err != nil || operationID != enqueued.OperationID.String() {
		return formscleanupapp.ErrUnavailable
	}
	return nil
}
