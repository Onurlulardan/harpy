package systemtransfer

import (
	"context"
	"time"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (storage *store) CancelExport(ctx context.Context, read app.ControlRead,
	now time.Time) (app.ExportRecord, error) {
	if !storage.usable(ctx) || !storage.command || read.Validate() != nil ||
		!controlTime(now) || read.Evidence.TenantID != storage.tenantID ||
		read.Evidence.ActorID != storage.actorID {
		return app.ExportRecord{}, app.ErrControlUnavailable
	}
	var result app.ExportRecord
	err := storage.queryRow(ctx, cancelExportSQL, read.TransferID, storage.actorID,
		string(read.Evidence.Resource), read.Evidence.TokenVersion,
		read.Evidence.AuthorizationVersion, now).Scan(&result.ExportID, &result.JobID,
		&result.Resource, &result.Format, &result.State, &result.Rows, &result.Bytes,
		&result.CreatedAt, &result.ReadyAt, &result.ExpiresAt)
	if err != nil {
		return app.ExportRecord{}, mapError(err, storage.notFound)
	}
	return result, nil
}
