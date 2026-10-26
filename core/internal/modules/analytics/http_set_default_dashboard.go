package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) SetDefaultDashboard(ctx context.Context, access analyticshttp.Access,
	request dashboardapp.SetDefaultRequest) error {
	return runtime.command(ctx, access, analyticshttp.PermissionDashboardRead,
		analyticshttp.OperationDashboardSetDefault, func(scoped context.Context, scope platformtx.TenantScope) error {
			return runtime.dashboards.SetDefault(scoped, scope, request)
		})
}
