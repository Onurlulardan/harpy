package catalog

import (
	"context"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) Restore(ctx context.Context, access cataloghttp.Access,
	installationID uuid.UUID, key string,
) (cataloghttp.Operation, error) {
	var result cataloghttp.Operation
	err := runtime.lifecycleCommand(ctx, access, cataloghttp.PermissionRestore,
		"catalog.installation.restore", func(scoped context.Context, scope platformtx.TenantScope) error {
			installation, err := runtime.lifecycle.CommandInstallation(
				scoped, scope, installationID, lifecycleapp.ActionRestore)
			if err != nil {
				return err
			}
			packageID, err := uuid.Parse(installation.PackageID.String())
			if err != nil {
				return cataloghttp.ErrUnavailable
			}
			handle, err := runtime.lifecycle.Restore(scoped, scope,
				lifecycleapp.PackageCommand{PackageID: packageID, IdempotencyKey: key})
			if err == nil {
				result = cataloghttp.RepresentHandle(handle)
			}
			return err
		})
	return result, err
}
