package analytics

import (
	"context"

	authority "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/authority"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type queryAuthorizer struct{}

func (queryAuthorizer) Authorize(
	ctx context.Context, scope platformtx.TenantScope, permission queryapp.Permission,
) error {
	if !analyticsAuthorityUsable(ctx, scope, string(permission)) {
		return queryapp.ErrForbidden
	}
	return nil
}

func analyticsAuthorityUsable(
	ctx context.Context, scope platformtx.TenantScope, permission string,
) bool {
	if ctx == nil || ctx.Err() != nil || platformtx.NetworkAllowed(ctx) ||
		!platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireOwner(scope, "analytics") != nil {
		return false
	}
	evidence, found := authority.FromContext(ctx)
	return found && evidence.Authorizes(scope, permission)
}

var _ queryapp.Authorizer = queryAuthorizer{}
