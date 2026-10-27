package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (storage *store) ExportRequest(ctx context.Context,
	write app.ExportWrite) (app.ExportRecord, error) {
	if !storage.usable(ctx) || !storage.command || write.Validate() != nil ||
		write.Evidence.TenantID != storage.tenantID ||
		write.Evidence.ActorID != storage.actorID {
		return app.ExportRecord{}, app.ErrControlUnavailable
	}
	var result app.ExportRecord
	err := storage.queryRow(ctx, submitExportSQL, write.ExportID, write.JobID,
		write.ReceiptID, storage.actorID, string(write.Evidence.Resource),
		write.Evidence.TokenVersion, write.Evidence.AuthorizationVersion,
		write.Evidence.AuthorizationClock, write.CatalogVersionID,
		write.CatalogSHA256[:], write.QueryDocument, write.QuerySHA256[:],
		write.RequestSHA256[:], write.IdempotencyKeySHA256[:], write.Format,
		write.CreatedAt).Scan(&result.ExportID, &result.JobID, &result.Resource,
		&result.Format, &result.State, &result.Rows, &result.Bytes,
		&result.CreatedAt, &result.ReadyAt, &result.ExpiresAt, &result.Replay)
	if err != nil {
		return app.ExportRecord{}, mapError(err, storage.notFound)
	}
	return result, nil
}
