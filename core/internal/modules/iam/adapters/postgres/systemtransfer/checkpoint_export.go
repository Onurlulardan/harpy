package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (store *exportWorkerBound) CheckpointExport(
	ctx context.Context,
	attempt app.ExportAttempt,
	progress app.ExportProgress,
) error {
	if !store.usable(ctx, attempt) || progress.Validate() != nil {
		return app.ErrExportWorkerInvalid
	}
	schema, err := store.schema(ctx)
	if err != nil {
		return err
	}
	arguments := append(outboundAttemptArguments(schema, attempt),
		progress.Rows, progress.PlaintextBytes, progress.CheckedAt)
	var replayed bool
	err = store.database.QueryRow(ctx, outboundCheckpointSQL,
		arguments...).Scan(&replayed)
	if err != nil {
		return mapOutboundError(err, app.ErrExportStateConflict)
	}
	if replayed {
		return app.ErrExportStateConflict
	}
	return nil
}
