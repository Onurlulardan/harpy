package restoreexecution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) CompleteRestoreExecution(
	ctx context.Context,
	record ports.RestoreCompletion,
) (bool, error) {
	return store.change(
		ctx, completeSQL, record.RestoreExecutionIdentity,
		record.ReceiptSHA256, record.ReceiptBytes,
	)
}
