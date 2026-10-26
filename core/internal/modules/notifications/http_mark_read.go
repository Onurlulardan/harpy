package notifications

import (
	"context"

	notificationhttp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) MarkRead(ctx context.Context,
	input notificationhttp.MessageInput,
) (result application.Message, err error) {
	values, err := runtime.resolve(ctx, input.Access,
		notificationhttp.PermissionInboxUpdate, notificationhttp.OperationInboxRead)
	if err != nil {
		return result, err
	}
	err = runtime.command(ctx, input.Access, values, notificationhttp.OperationInboxRead,
		func(operation context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.application.MarkRead(operation, scope, input.ID)
			return err
		})
	return result, httpPortError(err)
}
