package systemtransfer

import "context"

func ImportRequest(ctx context.Context, storage ControlStore,
	write ImportWrite) (ImportRecord, error) {
	if ctx == nil || ctx.Err() != nil || nilDependency(storage) ||
		write.Validate() != nil {
		return ImportRecord{}, ErrControlInvalid
	}
	return storage.ImportRequest(ctx, write)
}
