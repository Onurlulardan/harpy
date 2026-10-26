package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) bind(
	ctx context.Context,
	scope platformtx.TenantScope,
	access ScopeAccess,
	operation string,
	action Action,
) (Store, ScopeIdentity, error) {
	if application == nil || application.binder == nil || ctx == nil || ctx.Err() != nil {
		return nil, ScopeIdentity{}, ErrUnavailable
	}
	mode := platformtx.Command
	if access == ScopeSnapshot {
		mode = platformtx.SnapshotQuery
	}
	if !platformtx.ContextOwnsScope(ctx, scope) || scope.Mode() != mode || scope.OperationID() != operation {
		return nil, ScopeIdentity{}, ErrForbidden
	}
	store, identity, err := application.binder.Bind(ctx, scope, access)
	if err != nil || store == nil || identity.ActorID != scope.ActorID() ||
		identity.TenantID != scope.TenantID() {
		return nil, ScopeIdentity{}, ErrUnavailable
	}
	if err = application.authorizer.Authorize(ctx, scope, action); err != nil {
		return nil, ScopeIdentity{}, ErrForbidden
	}
	return store, identity, nil
}
