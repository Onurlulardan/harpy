package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) ListPublications(ctx context.Context, access analyticshttp.Access,
	limit int, after *uuid.UUID) (publicationapp.Page, error) {
	var result publicationapp.Page
	err := runtime.snapshot(ctx, access, analyticshttp.PermissionPublicationRead,
		analyticshttp.OperationPublicationList, func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.publications.List(scoped, scope, limit, after)
			return operationErr
		})
	return result, err
}
