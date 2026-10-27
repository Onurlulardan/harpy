package systemtransfer

import (
	"context"
	"time"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (storage *store) DownloadExportArtifact(ctx context.Context,
	read app.ControlRead, now time.Time) (app.ExportArtifact, error) {
	if !storage.usable(ctx) || read.Validate() != nil || !controlTime(now) ||
		read.Evidence.TenantID != storage.tenantID ||
		read.Evidence.ActorID != storage.actorID {
		return app.ExportArtifact{}, app.ErrControlUnavailable
	}
	var result app.ExportArtifact
	err := storage.queryRow(ctx, readExportArtifactSQL, read.TransferID,
		storage.actorID, string(read.Evidence.Resource), read.Evidence.TokenVersion,
		read.Evidence.AuthorizationVersion, now).Scan(&result.ExportID, &result.JobID,
		&result.Resource, &result.Format, &result.State, &result.Rows, &result.Bytes,
		&result.CreatedAt, &result.ReadyAt, &result.ExpiresAt, &result.ObjectKey,
		&result.ObjectVersion, &result.HeaderSHA256, &result.CiphertextSHA256,
		&result.CiphertextBytes, &result.KeyID, &result.WrappedDEK,
		&result.WrapNonce, &result.AttemptGeneration)
	if err != nil {
		return app.ExportArtifact{}, mapError(err, storage.notFound)
	}
	return result, nil
}

func controlTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%int(time.Microsecond) == 0
}
