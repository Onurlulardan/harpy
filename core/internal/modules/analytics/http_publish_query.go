package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) PublishQuery(ctx context.Context, access analyticshttp.Access,
	request queryapp.PublishRequest) (queryapp.Snapshot, error) {
	var result queryapp.Snapshot
	err := runtime.queryOwnerCommand(ctx, access, analyticshttp.PermissionQueryPublish,
		analyticshttp.OperationQueryPublish, func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.queries.Publish(scoped, scope, request)
			return operationErr
		})
	return result, err
}
