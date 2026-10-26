package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) UpdateQuery(ctx context.Context, access analyticshttp.Access,
	request queryapp.UpdateRequest) (queryapp.Snapshot, error) {
	var result queryapp.Snapshot
	err := runtime.command(ctx, access, analyticshttp.PermissionQueryUpdate,
		analyticshttp.OperationQueryUpdate, func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.queries.Update(scoped, scope, request)
			return operationErr
		})
	return result, err
}
