package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) PublishDashboard(ctx context.Context, access analyticshttp.Access,
	request dashboardapp.PublishRequest) (dashboardapp.Snapshot, error) {
	var result dashboardapp.Snapshot
	err := runtime.command(ctx, access, analyticshttp.PermissionDashboardPublish,
		analyticshttp.OperationDashboardPublish, func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.dashboards.Publish(scoped, scope, request)
			return operationErr
		})
	return result, err
}
