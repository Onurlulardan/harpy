package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) GetPublication(ctx context.Context, access analyticshttp.Access,
	releaseID uuid.UUID) (publicationapp.Snapshot, error) {
	var result publicationapp.Snapshot
	err := runtime.snapshot(ctx, access, analyticshttp.PermissionPublicationRead,
		analyticshttp.OperationPublicationGet, func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.publications.Get(scoped, scope, releaseID)
			return operationErr
		})
	return result, err
}
