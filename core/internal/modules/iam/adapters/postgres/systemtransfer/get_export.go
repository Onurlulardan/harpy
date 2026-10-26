package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (storage *store) GetExport(ctx context.Context,
	read app.ControlRead) (app.ExportRecord, error) {
	if !storage.usable(ctx) || read.Validate() != nil ||
		read.Evidence.TenantID != storage.tenantID ||
		read.Evidence.ActorID != storage.actorID {
		return app.ExportRecord{}, app.ErrControlUnavailable
	}
	var result app.ExportRecord
	err := storage.queryRow(ctx, readExportSQL, read.TransferID, storage.actorID,
		string(read.Evidence.Resource), read.Evidence.TokenVersion,
		read.Evidence.AuthorizationVersion).Scan(&result.ExportID, &result.JobID,
		&result.Resource, &result.Format, &result.State, &result.Rows, &result.Bytes,
		&result.CreatedAt, &result.ReadyAt, &result.ExpiresAt)
	if err != nil {
		return app.ExportRecord{}, mapError(err, storage.notFound)
	}
	return result, nil
}
