package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) ExecuteQuery(ctx context.Context, access analyticshttp.Access,
	request queryapp.ExecuteRequest) (queryapp.ExecutionResult, error) {
	var result queryapp.ExecutionResult
	err := runtime.queryOwnerSnapshot(ctx, access, analyticshttp.PermissionQueryExecute,
		analyticshttp.OperationQueryExecute, func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.queries.Execute(scoped, scope, request)
			return operationErr
		})
	return result, err
}
