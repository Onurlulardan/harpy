package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/google/uuid"
)

func CompileOwnerReceipts(
	ctx context.Context, store Store, operationID uuid.UUID,
) ([]catalogdomain.OwnerReceipt, error) {
	if ctx == nil || ctx.Err() != nil || store == nil {
		return nil, ErrUnavailable
	}
	return store.CompileOwnerReceipts(ctx, operationID)
}
