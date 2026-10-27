package registry

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) ResolveVersion(
	ctx context.Context, scope platformtx.TenantScope, packageID, version string,
) (packageregistry.SealedVersion, error) {
	registry, err := application.bind(ctx, scope, "catalog.package_version.get")
	if err != nil {
		return packageregistry.SealedVersion{}, err
	}
	return registry.ResolveVersion(ctx, packageID, version)
}
