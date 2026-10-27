package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
)

func RecordOwnerReceipt(
	ctx context.Context, store Store, receipt catalogdomain.OwnerReceipt,
) error {
	if ctx == nil || ctx.Err() != nil || store == nil {
		return ErrUnavailable
	}
	return store.RecordOwnerReceipt(ctx, receipt)
}
