package notifications

import (
	"context"

	notificationhttp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) UpdatePreferences(ctx context.Context,
	input notificationhttp.PreferencesInput,
) (result []application.Preference, err error) {
	values, err := runtime.resolve(ctx, input.Access,
		notificationhttp.PermissionPreferencesEdit, notificationhttp.OperationPreferencesEdit)
	if err != nil {
		return nil, err
	}
	err = runtime.command(ctx, input.Access, values, notificationhttp.OperationPreferencesEdit,
		func(operation context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.application.SetPreferences(operation, scope, input.Updates)
			return err
		})
	return result, httpPortError(err)
}
