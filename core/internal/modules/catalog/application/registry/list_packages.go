package registry

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) ListPackages(
	ctx context.Context, scope platformtx.TenantScope, limit int, after string,
) ([]packageregistry.Package, error) {
	registry, err := application.bind(ctx, scope, "catalog.package.list")
	if err != nil {
		return nil, err
	}
	return registry.ListPackages(ctx, limit, after)
}
