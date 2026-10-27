package analytics

import (
	"context"
	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) CreatePublication(ctx context.Context, access analyticshttp.Access,
	request publicationapp.CreateRequest) (publicationapp.Snapshot, error) {
	permission := analyticshttp.PermissionPublicationRead
	if request.IsPublic() {
		permission = analyticshttp.PermissionPublicationPublishPublic
	} else if request.IsCapability() {
		permission = analyticshttp.PermissionPublicationPublishCapability
	}
	var result publicationapp.Snapshot
	err := runtime.command(ctx, access, permission, analyticshttp.OperationPublicationCreate,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			result, operationErr = runtime.publications.Create(scoped, scope, request)
			return operationErr
		})
	return result, err
}
