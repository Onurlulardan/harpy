package item

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) bind(ctx context.Context, scope platformtx.TenantScope,
	access Access, operation, permission string,
) (Store, foundation.Identity, error) {
	if application == nil || ctx == nil || ctx.Err() != nil || scope.OperationID() != operation {
		return nil, foundation.Identity{}, ErrForbidden
	}
	if err := application.authorizer.Authorize(ctx, scope, permission); err != nil {
		return nil, foundation.Identity{}, ErrForbidden
	}
	store, identity, err := application.binder.Bind(ctx, scope, access)
	if err != nil || store == nil || !identity.Valid() {
		return nil, foundation.Identity{}, ErrUnavailable
	}
	return store, identity, nil
}
