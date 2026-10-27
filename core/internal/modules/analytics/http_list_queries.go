package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) ListQueries(ctx context.Context, access analyticshttp.Access,
	limit int, after *uuid.UUID) (queryapp.Page, error) {
	var result queryapp.Page
	err := runtime.snapshot(ctx, access, analyticshttp.PermissionQueryRead,
		analyticshttp.OperationQueryList, func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.queries.List(scoped, scope, limit, after)
			return operationErr
		})
	return result, err
}
