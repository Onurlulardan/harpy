package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (store *exportWorkerBound) AcquireOutbound(
	ctx context.Context,
	attempt app.ExportAttempt,
) (app.ExportClaimResult, error) {
	if !store.usable(ctx, attempt) {
		return app.ExportClaimResult{}, app.ErrExportWorkerInvalid
	}
	schema, err := store.schema(ctx)
	if err != nil {
		return app.ExportClaimResult{}, err
	}
	var state string
	var replayed bool
	err = store.database.QueryRow(ctx, outboundClaimSQL,
		outboundAttemptArguments(schema, attempt)...).Scan(&state, &replayed)
	if err != nil {
		return app.ExportClaimResult{}, mapOutboundError(
			err, app.ErrExportStateConflict)
	}
	result := app.ExportClaimResult{State: app.ExportOwnerState(state), Replay: replayed}
	if result.State != app.ExportRunning {
		return app.ExportClaimResult{}, app.ErrExportStateConflict
	}
	return result, nil
}
