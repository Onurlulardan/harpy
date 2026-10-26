package systemtransfer

import "context"

func ResolveImport(ctx context.Context, storage ControlStore,
	read ControlRead) (ImportRecord, error) {
	if ctx == nil || ctx.Err() != nil || nilDependency(storage) ||
		read.Validate() != nil {
		return ImportRecord{}, ErrControlInvalid
	}
	return storage.ResolveImport(ctx, read)
}
