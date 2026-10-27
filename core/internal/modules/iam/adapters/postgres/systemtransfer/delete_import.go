package systemtransfer

import (
	"context"
	"time"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (storage *store) DeleteImport(ctx context.Context, read app.ControlRead,
	now time.Time) (app.ImportRecord, error) {
	if !storage.usable(ctx) || !storage.command || read.Validate() != nil ||
		!controlTime(now) || read.Evidence.TenantID != storage.tenantID ||
		read.Evidence.ActorID != storage.actorID {
		return app.ImportRecord{}, app.ErrControlUnavailable
	}
	var result app.ImportRecord
	err := storage.queryRow(ctx, cancelImportSQL, read.TransferID, storage.actorID,
		string(read.Evidence.Resource), read.Evidence.TokenVersion,
		read.Evidence.AuthorizationVersion, now).Scan(&result.ImportID, &result.JobID,
		&result.Resource, &result.State, &result.RowsSeen, &result.Created,
		&result.Updated, &result.Skipped, &result.Failed, &result.CreatedAt,
		&result.CompletedAt)
	if err != nil {
		return app.ImportRecord{}, mapError(err, storage.notFound)
	}
	return result, nil
}
