package catalog

import (
	"context"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) ListPackageVersionDependencies(
	ctx context.Context, access cataloghttp.Access, id uuid.UUID, version string,
) ([]cataloghttp.PackageDependency, error) {
	var result []cataloghttp.PackageDependency
	err := runtime.snapshot(ctx, access, cataloghttp.PermissionRead,
		"catalog.package_version_dependency.list",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			values, err := runtime.dependency.List(scoped, scope, id.String(), version)
			if err != nil {
				return err
			}
			result = make([]cataloghttp.PackageDependency, 0, len(values))
			for _, value := range values {
				result = append(result, cataloghttp.RepresentDependency(value))
			}
			return nil
		})
	return result, err
}
