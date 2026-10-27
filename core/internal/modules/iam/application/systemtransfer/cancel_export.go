package systemtransfer

import (
	"context"
	"time"
)

func CancelExport(ctx context.Context, storage ControlStore, read ControlRead,
	now time.Time) (ExportRecord, error) {
	if ctx == nil || ctx.Err() != nil || nilDependency(storage) ||
		read.Validate() != nil || !controlTime(now) {
		return ExportRecord{}, ErrControlInvalid
	}
	return storage.CancelExport(ctx, read, now)
}
