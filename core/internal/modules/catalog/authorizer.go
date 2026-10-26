package catalog

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/authority"
	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type applicationAuthorizer struct{}

func (applicationAuthorizer) Authorize(
	ctx context.Context, scope platformtx.TenantScope, permission string,
) error {
	if !catalogAuthorized(ctx, scope, permission) {
		return errCatalogForbidden
	}
	return nil
}

type lifecycleAuthorizer struct{}

func (lifecycleAuthorizer) Authorize(ctx context.Context, scope platformtx.TenantScope,
	action lifecycleapp.Action,
) error {
	if !catalogAuthorized(ctx, scope, string(action)) {
		return lifecycleapp.ErrForbidden
	}
	return nil
}

func catalogAuthorized(ctx context.Context, scope platformtx.TenantScope, permission string) bool {
	if ctx == nil || ctx.Err() != nil || platformtx.NetworkAllowed(ctx) ||
		!platformtx.ContextOwnsScope(ctx, scope) || platformtx.RequireOwner(scope, "catalog") != nil {
		return false
	}
	evidence, found := authority.FromContext(ctx)
	return found && evidence.Authorizes(scope, permission)
}

type catalogAuthorizationError struct{}

func (catalogAuthorizationError) Error() string { return "catalog operation is forbidden" }

var errCatalogForbidden error = catalogAuthorizationError{}
