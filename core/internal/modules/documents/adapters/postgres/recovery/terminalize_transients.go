package recovery

import (
	"context"

	recoveryapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/recovery"
	"github.com/google/uuid"
)

func (store *Store) TerminalizeTransients(
	ctx context.Context,
	restoreOperationID uuid.UUID,
) (recoveryapp.TransientDisposition, error) {
	var result recoveryapp.TransientDisposition
	err := store.with(ctx, transactionCommand,
		func(database recoveryDatabase) error {
			return database.QueryRow(
				ctx, terminalizeTransientsSQL, restoreOperationID,
			).Scan(&result.CancelledUploads, &result.ExpiredArchives,
				&result.CancelledArchives)
		})
	return result, err
}
