package dashboardexport

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) bind(
	ctx context.Context, scope platformtx.TenantScope, access ScopeAccess, operation string,
) (Store, ScopeIdentity, error) {
	if application == nil || ctx == nil || ctx.Err() != nil || scope.OperationID() != operation {
		return nil, ScopeIdentity{}, ErrForbidden
	}
	if err := application.authorizer.Authorize(ctx, scope, PermissionExport); err != nil {
		return nil, ScopeIdentity{}, ErrForbidden
	}
	store, identity, err := application.binder.Bind(ctx, scope, access)
	if err != nil || store == nil || identity.ActorID == uuid.Nil || identity.TenantID == uuid.Nil ||
		identity.ActorID != scope.ActorID() || identity.TenantID != scope.TenantID() {
		return nil, ScopeIdentity{}, ErrUnavailable
	}
	return store, identity, nil
}
