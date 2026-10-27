package notifications

import (
	"context"

	notificationhttp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) GetMessage(ctx context.Context,
	input notificationhttp.MessageInput,
) (result application.Message, err error) {
	values, err := runtime.resolve(ctx, input.Access,
		notificationhttp.PermissionInboxRead, notificationhttp.OperationInboxGet)
	if err != nil {
		return result, err
	}
	err = runtime.snapshot(ctx, input.Access, values, notificationhttp.OperationInboxGet,
		func(operation context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.application.GetMessage(operation, scope, input.ID)
			return err
		})
	return result, httpPortError(err)
}
