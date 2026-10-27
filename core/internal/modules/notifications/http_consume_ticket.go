package notifications

import (
	"context"

	notificationhttp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) ConsumeRealtimeTicket(ctx context.Context,
	input notificationhttp.RealtimeInput,
) (result application.RealtimeConnectionIdentity, err error) {
	operation := notificationhttp.OperationSocketConnect
	if input.Legacy {
		operation = notificationhttp.OperationLegacyConnect
	}
	values, err := runtime.resolve(ctx, input.Access,
		notificationhttp.PermissionInboxRead, operation)
	if err != nil {
		return result, err
	}
	err = runtime.command(ctx, input.Access, values, operation,
		func(operationContext context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.realtime.ConsumeTicket(operationContext, scope,
				input.Access.Origin(), input.OfferedProtocols)
			if err == nil && input.Legacy {
				err = runtime.realtime.SignalLegacyConnection(
					operationContext, scope, input.Access.ConsumerManifestDigest())
			}
			return err
		})
	return result, httpPortError(err)
}
