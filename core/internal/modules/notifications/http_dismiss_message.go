package notifications

import (
	"context"

	notificationhttp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) DismissMessage(ctx context.Context,
	input notificationhttp.MessageInput,
) (result application.Message, err error) {
	values, err := runtime.resolve(ctx, input.Access,
		notificationhttp.PermissionInboxDelete, notificationhttp.OperationInboxDismiss)
	if err != nil {
		return result, err
	}
	err = runtime.command(ctx, input.Access, values, notificationhttp.OperationInboxDismiss,
		func(operation context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.application.DismissMessage(operation, scope, input.ID)
			return err
		})
	return result, httpPortError(err)
}
