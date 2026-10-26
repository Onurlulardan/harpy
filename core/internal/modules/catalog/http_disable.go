package catalog

import (
	"context"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) Disable(ctx context.Context, access cataloghttp.Access,
	installationID uuid.UUID, key string,
) (cataloghttp.Installation, error) {
	var result cataloghttp.Installation
	err := runtime.lifecycleCommand(ctx, access, cataloghttp.PermissionDisable,
		"catalog.installation.disable", func(scoped context.Context, scope platformtx.TenantScope) error {
			installation, err := runtime.lifecycle.CommandInstallation(
				scoped, scope, installationID, lifecycleapp.ActionDisable)
			if err != nil {
				return err
			}
			packageID, err := uuid.Parse(installation.PackageID.String())
			if err != nil {
				return cataloghttp.ErrUnavailable
			}
			value, err := runtime.lifecycle.Disable(scoped, scope,
				lifecycleapp.PackageCommand{PackageID: packageID, IdempotencyKey: key})
			if err == nil {
				result = cataloghttp.RepresentInstallation(value)
			}
			return err
		})
	return result, err
}
