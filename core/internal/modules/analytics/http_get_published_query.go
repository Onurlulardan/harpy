package analytics

import (
	"context"

	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) GetPublishedQuery(
	ctx context.Context, access analyticshttp.Access, queryID uuid.UUID,
) (queryapp.PublishedSnapshot, error) {
	var result queryapp.PublishedSnapshot
	err := runtime.snapshot(
		ctx,
		access,
		analyticshttp.PermissionQueryRead,
		analyticshttp.OperationQueryGetPublished,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.queries.ReadPublished(
				scoped,
				scope,
				queryID,
			)
			return operationErr
		},
	)
	return result, err
}
