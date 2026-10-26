package systemtransfer

import (
	"context"
	"time"
)

func DeleteImport(ctx context.Context, storage ControlStore, read ControlRead,
	now time.Time) (ImportRecord, error) {
	if ctx == nil || ctx.Err() != nil || nilDependency(storage) ||
		read.Validate() != nil || !controlTime(now) {
		return ImportRecord{}, ErrControlInvalid
	}
	return storage.DeleteImport(ctx, read, now)
}
