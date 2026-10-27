package notifications

import (
	"context"

	notificationhttp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/http"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) MarkAllRead(ctx context.Context,
	access notificationhttp.Access,
) (result int64, err error) {
	values, err := runtime.resolve(ctx, access,
		notificationhttp.PermissionInboxUpdate, notificationhttp.OperationInboxReadAll)
	if err != nil {
		return 0, err
	}
	err = runtime.command(ctx, access, values, notificationhttp.OperationInboxReadAll,
		func(operation context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.application.CompleteInboxRead(operation, scope)
			return err
		})
	return result, httpPortError(err)
}
