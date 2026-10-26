package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (storage *store) ImportRequest(ctx context.Context,
	write app.ImportWrite) (app.ImportRecord, error) {
	if !storage.usable(ctx) || !storage.command || write.Validate() != nil ||
		write.Evidence.TenantID != storage.tenantID ||
		write.Evidence.ActorID != storage.actorID {
		return app.ImportRecord{}, app.ErrControlUnavailable
	}
	var result app.ImportRecord
	err := storage.queryRow(ctx, submitImportSQL, write.ImportID, write.JobID,
		write.ReceiptID, storage.actorID, string(write.Evidence.Resource), write.SourceID,
		write.Evidence.TokenVersion, write.Evidence.AuthorizationVersion,
		write.Evidence.AuthorizationClock, write.CatalogVersionID,
		write.CatalogSHA256[:], write.IdentityManifestSHA256[:],
		write.RequestSHA256[:], write.IdempotencyKeySHA256[:], write.DuplicateAction,
		write.MappingDocument, write.MappingSHA256[:], write.FieldCount,
		write.CreatedAt).Scan(&result.ImportID, &result.JobID, &result.Resource,
		&result.State, &result.RowsSeen, &result.Created, &result.Updated,
		&result.Skipped, &result.Failed, &result.CreatedAt,
		&result.CompletedAt, &result.Replay)
	if err != nil {
		return app.ImportRecord{}, mapError(err, storage.notFound)
	}
	return result, nil
}
