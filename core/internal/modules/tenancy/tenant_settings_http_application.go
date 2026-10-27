package tenancy

import (
	"context"

	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *TenantSettingsHTTPRuntime) Get(ctx context.Context,
	access TenantSettingsAccess,
) (result settings.Snapshot, err error) {
	err = runtime.snapshot(ctx, access, TenantSettingsOperationGet,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.application.Get(scoped, scope)
			return err
		})
	return result, err
}

func (runtime *TenantSettingsHTTPRuntime) Update(ctx context.Context,
	access TenantSettingsAccess, command settings.UpdateCommand,
) (result settings.UpdateResult, err error) {
	err = runtime.command(ctx, access, TenantSettingsOperationUpdate, command.SMTP != nil,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.application.Update(scoped, scope, command)
			return err
		})
	return result, err
}

func (runtime *TenantSettingsHTTPRuntime) TestEmail(ctx context.Context,
	access TenantSettingsAccess, command settings.TestEmailCommand,
) (result settings.TestEmailResult, err error) {
	err = runtime.command(ctx, access, TenantSettingsOperationTestEmail, true,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.application.TestEmail(scoped, scope, command)
			return err
		})
	return result, err
}
