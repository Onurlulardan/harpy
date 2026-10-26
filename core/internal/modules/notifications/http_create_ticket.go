package notifications

import (
	"context"

	notificationhttp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) CreateRealtimeTicket(ctx context.Context,
	access notificationhttp.Access,
) (result application.RealtimeTicketResult, err error) {
	values, err := runtime.resolve(ctx, access,
		notificationhttp.PermissionInboxRead, notificationhttp.OperationTicketCreate)
	if err != nil {
		return result, err
	}
	err = runtime.command(ctx, access, values, notificationhttp.OperationTicketCreate,
		func(operation context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.realtime.IssueTicket(operation, scope, access.Origin())
			return err
		})
	return result, httpPortError(err)
}
