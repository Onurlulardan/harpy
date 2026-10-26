package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) ActivatePublication(ctx context.Context, access analyticshttp.Access,
	request publicationapp.ActivateRequest) (publicationapp.Activation, error) {
	var result publicationapp.Activation
	err := runtime.command(ctx, access, analyticshttp.PermissionPublicationRead,
		analyticshttp.OperationPublicationActivate, func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.publications.Activate(scoped, scope, request)
			return operationErr
		})
	return result, err
}
