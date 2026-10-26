package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) CreateDashboard(ctx context.Context, access analyticshttp.Access,
	request dashboardapp.CreateRequest) (dashboardapp.Snapshot, error) {
	var result dashboardapp.Snapshot
	err := runtime.command(ctx, access, analyticshttp.PermissionDashboardCreate,
		analyticshttp.OperationDashboardCreate, func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.dashboards.Create(scoped, scope, request)
			return operationErr
		})
	return result, err
}
