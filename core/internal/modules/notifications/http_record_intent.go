package notifications

import (
	"context"

	notificationhttp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) RecordIntent(ctx context.Context,
	input notificationhttp.IntentInput,
) (result api.IntentResult, err error) {
	values, err := runtime.resolve(ctx, input.Access,
		notificationhttp.PermissionSend, notificationhttp.OperationIntentCreate)
	if err != nil {
		return result, err
	}
	err = runtime.command(ctx, input.Access, values, notificationhttp.OperationIntentCreate,
		func(operation context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.application.RecordIntent(operation, scope, input.Request)
			return err
		})
	return result, httpPortError(err)
}
