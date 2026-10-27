package notifications

import (
	"context"

	notificationhttp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) GetPreferences(ctx context.Context,
	access notificationhttp.Access,
) (result []application.Preference, err error) {
	values, err := runtime.resolve(ctx, access,
		notificationhttp.PermissionPreferencesRead, notificationhttp.OperationPreferencesGet)
	if err != nil {
		return nil, err
	}
	err = runtime.snapshot(ctx, access, values, notificationhttp.OperationPreferencesGet,
		func(operation context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.application.ReadPreferences(operation, scope)
			return err
		})
	return result, httpPortError(err)
}
