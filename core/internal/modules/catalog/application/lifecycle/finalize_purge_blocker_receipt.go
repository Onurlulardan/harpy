package lifecycle

import "context"

func FinalizePurgeBlockerReceipt(
	ctx context.Context, store Store, receipt PurgeBlockerReceipt,
) error {
	if ctx == nil || ctx.Err() != nil || store == nil {
		return ErrUnavailable
	}
	return store.FinalizePurgeBlockerReceipt(ctx, receipt)
}
