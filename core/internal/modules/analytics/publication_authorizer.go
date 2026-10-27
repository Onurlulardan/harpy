package analytics

import (
	"context"
	"time"

	authority "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/authority"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type publicationAuthorizer struct{ principals dashboardPrincipalAuthorizer }

func (authorizer publicationAuthorizer) Authorize(
	ctx context.Context, scope platformtx.TenantScope, permission publicationapp.Permission,
) error {
	if analyticsAuthorityUsable(ctx, scope, string(permission)) {
		return nil
	}
	if authorizer.principals == nil ||
		authorizer.principals.AuthorizePrincipal(
			ctx, scope, "user", scope.ActorID(), string(permission),
		) != nil {
		return publicationapp.ErrForbidden
	}
	return nil
}

func (publicationAuthorizer) FreshStepUp(
	ctx context.Context, scope platformtx.TenantScope, now time.Time,
) (time.Time, error) {
	if ctx == nil || ctx.Err() != nil || !platformtx.ContextOwnsScope(ctx, scope) {
		return time.Time{}, publicationapp.ErrForbidden
	}
	evidence, found := authority.FromContext(ctx)
	stepUp, fresh := evidence.FreshStepUp(now.UTC().Truncate(time.Second), 10*time.Minute)
	if !found || !fresh {
		return time.Time{}, publicationapp.ErrForbidden
	}
	return stepUp, nil
}

var _ publicationapp.Authorizer = publicationAuthorizer{}
