package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) DeleteQuery(ctx context.Context, access analyticshttp.Access,
	request queryapp.DeleteRequest) (queryapp.Snapshot, error) {
	var result queryapp.Snapshot
	err := runtime.queryOwnerCommand(ctx, access, analyticshttp.PermissionQueryDelete,
		analyticshttp.OperationQueryDelete, func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.queries.Delete(scoped, scope, request)
			return operationErr
		})
	return result, err
}
