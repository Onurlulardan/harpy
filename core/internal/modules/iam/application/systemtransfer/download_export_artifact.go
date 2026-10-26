package systemtransfer

import (
	"context"
	"time"
)

func DownloadExportArtifact(ctx context.Context, storage ControlStore,
	read ControlRead, now time.Time) (ExportArtifact, error) {
	if ctx == nil || ctx.Err() != nil || nilDependency(storage) ||
		read.Validate() != nil || !controlTime(now) {
		return ExportArtifact{}, ErrControlInvalid
	}
	return storage.DownloadExportArtifact(ctx, read, now)
}
