package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) GetQuery(ctx context.Context, access analyticshttp.Access,
	queryID uuid.UUID) (queryapp.Snapshot, error) {
	var result queryapp.Snapshot
	err := runtime.snapshot(ctx, access, analyticshttp.PermissionQueryRead,
		analyticshttp.OperationQueryGet, func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.queries.Get(scoped, scope, queryID)
			return operationErr
		})
	return result, err
}
