package systemtransfer

import "context"

func ListReceipts(ctx context.Context, storage ControlStore, read ControlRead,
	after int64, limit int) (ReceiptPage, error) {
	if ctx == nil || ctx.Err() != nil || nilDependency(storage) ||
		read.Validate() != nil || after < 0 || after > 1_000_000 ||
		limit < 1 || limit > 100 {
		return ReceiptPage{}, ErrControlInvalid
	}
	return storage.ListReceipts(ctx, read, after, limit)
}
