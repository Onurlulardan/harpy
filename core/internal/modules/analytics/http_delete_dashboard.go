package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) DeleteDashboard(ctx context.Context, access analyticshttp.Access,
	request dashboardapp.DeleteRequest) (dashboardapp.Snapshot, error) {
	var result dashboardapp.Snapshot
	err := runtime.command(ctx, access, analyticshttp.PermissionDashboardDelete,
		analyticshttp.OperationDashboardDelete, func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.dashboards.Delete(scoped, scope, request)
			return operationErr
		})
	return result, err
}
