package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) CreateQuery(ctx context.Context, access analyticshttp.Access,
	request queryapp.CreateRequest) (queryapp.Snapshot, error) {
	var result queryapp.Snapshot
	err := runtime.command(ctx, access, analyticshttp.PermissionQueryCreate,
		analyticshttp.OperationQueryCreate, func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.queries.Create(scoped, scope, request)
			return operationErr
		})
	return result, err
}
