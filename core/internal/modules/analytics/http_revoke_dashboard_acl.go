package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) RevokeDashboardACL(ctx context.Context, access analyticshttp.Access,
	request dashboardapp.RevokeACLRequest) error {
	return runtime.command(ctx, access, analyticshttp.PermissionDashboardUpdate,
		analyticshttp.OperationDashboardRevokeACL, func(scoped context.Context, scope platformtx.TenantScope) error {
			return runtime.dashboards.RevokeACL(scoped, scope, request)
		})
}
