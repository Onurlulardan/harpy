package notifications

import (
	"context"

	notificationhttp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) ListMessages(ctx context.Context,
	input notificationhttp.ListInput,
) (result ports.MessagePage, err error) {
	values, err := runtime.resolve(ctx, input.Access,
		notificationhttp.PermissionInboxRead, notificationhttp.OperationInboxList)
	if err != nil {
		return result, err
	}
	err = runtime.snapshot(ctx, input.Access, values, notificationhttp.OperationInboxList,
		func(operation context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.application.ListMessages(operation, scope, input.Filter)
			return err
		})
	return result, httpPortError(err)
}
