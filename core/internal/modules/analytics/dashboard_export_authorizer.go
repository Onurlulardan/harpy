package analytics

import (
	"context"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type dashboardExportAuthorizer struct{}

func (dashboardExportAuthorizer) Authorize(
	ctx context.Context, scope platformtx.TenantScope, permission exportapp.Permission,
) error {
	if !analyticsAuthorityUsable(ctx, scope, string(permission)) {
		return exportapp.ErrForbidden
	}
	return nil
}

var _ exportapp.Authorizer = dashboardExportAuthorizer{}
