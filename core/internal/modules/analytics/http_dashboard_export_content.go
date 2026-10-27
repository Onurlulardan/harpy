package analytics

import (
	"context"

	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) OpenDashboardExportContent(
	ctx context.Context, access analyticshttp.Access, dashboardID, exportID uuid.UUID,
) (exportapp.Content, error) {
	var stored exportapp.StoredContent
	err := runtime.snapshot(ctx, access, analyticshttp.PermissionDashboardExport,
		analyticshttp.OperationDashboardExportContent,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			stored, operationErr = runtime.exports.DescribeContent(
				scoped, scope, dashboardID, exportID)
			return operationErr
		})
	if err != nil {
		return exportapp.Content{}, err
	}
	return runtime.exports.OpenContent(ctx, stored)
}
