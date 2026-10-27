package analytics

import (
	"context"

	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) GetDashboardExport(
	ctx context.Context, access analyticshttp.Access, dashboardID, exportID uuid.UUID,
) (exportapp.Snapshot, error) {
	var result exportapp.Snapshot
	err := runtime.snapshot(ctx, access, analyticshttp.PermissionDashboardExport,
		analyticshttp.OperationDashboardExportGet,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.exports.Get(scoped, scope, dashboardID, exportID)
			return operationErr
		})
	return result, err
}
