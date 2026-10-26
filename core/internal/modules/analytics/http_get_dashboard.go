package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) GetDashboard(ctx context.Context, access analyticshttp.Access,
	dashboardID uuid.UUID) (dashboardapp.Snapshot, error) {
	var result dashboardapp.Snapshot
	err := runtime.snapshot(ctx, access, analyticshttp.PermissionDashboardRead,
		analyticshttp.OperationDashboardGet, func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.dashboards.Get(scoped, scope, dashboardID)
			return operationErr
		})
	return result, err
}
