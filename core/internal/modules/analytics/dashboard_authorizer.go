package analytics

import (
	"context"

	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type dashboardPrincipalAuthorizer interface {
	AuthorizePrincipal(context.Context, platformtx.TenantScope,
		string, uuid.UUID, string) error
}

type dashboardAuthorizer struct{ principals dashboardPrincipalAuthorizer }

func (dashboardAuthorizer) Authorize(
	ctx context.Context, scope platformtx.TenantScope, permission dashboardapp.Permission,
) error {
	if !analyticsAuthorityUsable(ctx, scope, string(permission)) {
		return dashboardapp.ErrForbidden
	}
	return nil
}

func (authorizer dashboardAuthorizer) AuthorizePrincipal(
	ctx context.Context, scope platformtx.TenantScope,
	principalKind string, principalID uuid.UUID, permission dashboardapp.Permission,
) error {
	if !analyticsAuthorityUsable(ctx, scope, string(dashboardapp.PermissionUpdate)) ||
		authorizer.principals == nil || principalID == uuid.Nil ||
		(principalKind != "user" && principalKind != "role") ||
		authorizer.principals.AuthorizePrincipal(
			ctx, scope, principalKind, principalID, string(permission),
		) != nil {
		return dashboardapp.ErrForbidden
	}
	return nil
}

var _ dashboardapp.Authorizer = dashboardAuthorizer{}
