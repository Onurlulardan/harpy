package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) UpdateDashboard(ctx context.Context, access analyticshttp.Access,
	request dashboardapp.UpdateRequest) (dashboardapp.Snapshot, error) {
	var result dashboardapp.Snapshot
	err := runtime.command(ctx, access, analyticshttp.PermissionDashboardUpdate,
		analyticshttp.OperationDashboardUpdate, func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.dashboards.Update(scoped, scope, request)
			return operationErr
		})
	return result, err
}
