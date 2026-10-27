package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) SubmitPublication(ctx context.Context, access analyticshttp.Access,
	request publicationapp.TransitionRequest) (publicationapp.Snapshot, error) {
	var result publicationapp.Snapshot
	err := runtime.command(ctx, access, analyticshttp.PermissionPublicationRead,
		analyticshttp.OperationPublicationSubmit, func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.publications.Submit(scoped, scope, request)
			return operationErr
		})
	return result, err
}
