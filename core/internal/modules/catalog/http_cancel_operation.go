package catalog

import (
	"context"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) CancelOperation(ctx context.Context, access cataloghttp.Access,
	operationID uuid.UUID, key string,
) (cataloghttp.Operation, error) {
	var result cataloghttp.Operation
	err := runtime.lifecycleCommand(ctx, access, cataloghttp.PermissionInstall,
		"catalog.operation.cancel", func(scoped context.Context, scope platformtx.TenantScope) error {
			value, err := runtime.lifecycle.Cancel(scoped, scope, lifecycleapp.CancelRequest{
				OperationID: operationID, IdempotencyKey: key})
			if err == nil {
				result = cataloghttp.RepresentOperation(value, false)
			}
			return err
		})
	return result, err
}
