package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) GrantDashboardACL(ctx context.Context, access analyticshttp.Access,
	request dashboardapp.GrantACLRequest) (dashboardapp.ACL, error) {
	var result dashboardapp.ACL
	err := runtime.command(ctx, access, analyticshttp.PermissionDashboardUpdate,
		analyticshttp.OperationDashboardGrantACL, func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.dashboards.GrantACL(scoped, scope, request)
			return operationErr
		})
	return result, err
}
