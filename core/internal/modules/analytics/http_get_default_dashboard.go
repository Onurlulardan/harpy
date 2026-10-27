package analytics

import (
	"context"

	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) GetDefaultDashboard(
	ctx context.Context, access analyticshttp.Access, surface string,
) (dashboardapp.Snapshot, error) {
	var result dashboardapp.Snapshot
	err := runtime.snapshot(ctx, access, analyticshttp.PermissionDashboardRead,
		analyticshttp.OperationDashboardRead,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.dashboards.ReadDefault(scoped, scope, surface)
			return operationErr
		})
	return result, err
}
