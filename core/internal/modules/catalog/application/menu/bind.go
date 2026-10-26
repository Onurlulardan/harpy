package menu

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) bind(
	ctx context.Context, scope platformtx.TenantScope, access ScopeAccess,
	operation, permission string,
) (Store, Identity, error) {
	if application == nil || ctx == nil || ctx.Err() != nil {
		return nil, Identity{}, ErrUnavailable
	}
	mode := platformtx.Command
	if access == ScopeSnapshot {
		mode = platformtx.SnapshotQuery
	}
	if !platformtx.ContextOwnsScope(ctx, scope) || scope.Mode() != mode || scope.OperationID() != operation {
		return nil, Identity{}, ErrForbidden
	}
	store, identity, err := application.binder.Bind(ctx, scope, access)
	if err != nil || store == nil || identity.ActorID != scope.ActorID() || identity.TenantID != scope.TenantID() {
		return nil, Identity{}, ErrUnavailable
	}
	if err = application.authorizer.Authorize(ctx, scope, permission); err != nil {
		return nil, Identity{}, ErrForbidden
	}
	return store, identity, nil
}
