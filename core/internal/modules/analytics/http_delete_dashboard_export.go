package analytics

import (
	"context"

	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) DeleteDashboardExport(
	ctx context.Context, access analyticshttp.Access, request exportapp.DeleteRequest,
) (exportapp.Snapshot, error) {
	var result exportapp.Snapshot
	err := runtime.command(ctx, access, analyticshttp.PermissionDashboardExport,
		analyticshttp.OperationDashboardExportDelete,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.exports.Delete(scoped, scope, request)
			return operationErr
		})
	return result, err
}
