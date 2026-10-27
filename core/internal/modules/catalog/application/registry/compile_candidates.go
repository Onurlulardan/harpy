package registry

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) CompileCandidates(
	ctx context.Context, scope platformtx.TenantScope, packageID string, limit int,
) ([]packageregistry.SealedVersion, error) {
	registry, err := application.bind(ctx, scope, "catalog.package_version.list")
	if err != nil {
		return nil, err
	}
	return registry.CompileCandidates(ctx, packageID, limit)
}
