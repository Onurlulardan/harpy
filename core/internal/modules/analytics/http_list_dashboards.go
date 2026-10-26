package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) ListDashboards(ctx context.Context, access analyticshttp.Access,
	limit int, after *uuid.UUID) (dashboardapp.Page, error) {
	var result dashboardapp.Page
	err := runtime.snapshot(ctx, access, analyticshttp.PermissionDashboardRead,
		analyticshttp.OperationDashboardList, func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.dashboards.List(scoped, scope, limit, after)
			return operationErr
		})
	return result, err
}
