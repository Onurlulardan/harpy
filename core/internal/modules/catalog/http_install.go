package catalog

import (
	"context"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) Install(ctx context.Context, access cataloghttp.Access,
	packageID uuid.UUID, version, key string,
) (cataloghttp.Operation, error) {
	var result cataloghttp.Operation
	err := runtime.lifecycleCommand(ctx, access, cataloghttp.PermissionInstall,
		"catalog.installation.install", func(scoped context.Context, scope platformtx.TenantScope) error {
			handle, err := runtime.lifecycle.Install(scoped, scope, lifecycleapp.InstallRequest{
				PackageID: packageID, Version: version, IdempotencyKey: key})
			if err == nil {
				result = cataloghttp.RepresentHandle(handle)
			}
			return err
		})
	return result, err
}
