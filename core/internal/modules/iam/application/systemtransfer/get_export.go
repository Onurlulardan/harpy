package systemtransfer

import "context"

func GetExport(ctx context.Context, storage ControlStore,
	read ControlRead) (ExportRecord, error) {
	if ctx == nil || ctx.Err() != nil || nilDependency(storage) ||
		read.Validate() != nil {
		return ExportRecord{}, ErrControlInvalid
	}
	return storage.GetExport(ctx, read)
}
