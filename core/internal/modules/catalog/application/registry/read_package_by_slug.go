package registry

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) ReadPackageBySlug(
	ctx context.Context, scope platformtx.TenantScope, slug string,
) (packageregistry.Package, error) {
	registry, err := application.bind(ctx, scope, "catalog.package.get")
	if err != nil {
		return packageregistry.Package{}, err
	}
	return registry.ReadPackageBySlug(ctx, slug)
}
