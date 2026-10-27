package item

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) mutationScope(ctx context.Context,
	scope platformtx.TenantScope, request MutationRequest, operation, permission string,
) (Store, foundation.Identity, error) {
	if !request.ItemID.UUIDv7() || request.RowVersion == 0 ||
		!foundation.ValidIdempotency(request.IdempotencyKey) {
		return nil, foundation.Identity{}, ErrInvalid
	}
	return application.bind(ctx, scope, Command, operation, permission)
}
