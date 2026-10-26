package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (storage *store) ResolveImport(ctx context.Context,
	read app.ControlRead) (app.ImportRecord, error) {
	if !storage.usable(ctx) || read.Validate() != nil ||
		read.Evidence.TenantID != storage.tenantID ||
		read.Evidence.ActorID != storage.actorID {
		return app.ImportRecord{}, app.ErrControlUnavailable
	}
	var result app.ImportRecord
	err := storage.queryRow(ctx, readImportSQL, read.TransferID, storage.actorID,
		string(read.Evidence.Resource), read.Evidence.TokenVersion,
		read.Evidence.AuthorizationVersion).Scan(&result.ImportID, &result.JobID,
		&result.Resource, &result.State, &result.RowsSeen, &result.Created,
		&result.Updated, &result.Skipped, &result.Failed, &result.CreatedAt,
		&result.CompletedAt)
	if err != nil {
		return app.ImportRecord{}, mapError(err, storage.notFound)
	}
	return result, nil
}
