package systemtransfer

import "context"

func ExportRequest(ctx context.Context, storage ControlStore,
	write ExportWrite) (ExportRecord, error) {
	if ctx == nil || ctx.Err() != nil || nilDependency(storage) ||
		write.Validate() != nil {
		return ExportRecord{}, ErrControlInvalid
	}
	return storage.ExportRequest(ctx, write)
}
