package catalog

import (
	"context"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) ListPackages(ctx context.Context, access cataloghttp.Access,
	limit int, after *uuid.UUID,
) ([]cataloghttp.Package, *uuid.UUID, error) {
	var values []cataloghttp.Package
	var next *uuid.UUID
	err := runtime.snapshot(ctx, access, cataloghttp.PermissionRead, "catalog.package.list",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			cursor := ""
			if after != nil {
				cursor = after.String()
			}
			items, err := runtime.registry.ListPackages(scoped, scope, limit, cursor)
			if err != nil {
				return err
			}
			values = make([]cataloghttp.Package, 0, len(items))
			for _, item := range items {
				values = append(values, cataloghttp.RepresentPackage(item))
			}
			if len(items) == limit && len(items) > 0 {
				parsed, parseErr := uuid.Parse(items[len(items)-1].ID)
				if parseErr != nil {
					return cataloghttp.ErrUnavailable
				}
				next = &parsed
			}
			return nil
		})
	return values, next, err
}

func (runtime *HTTPRuntime) GetPackage(ctx context.Context, access cataloghttp.Access,
	id uuid.UUID,
) (cataloghttp.Package, error) {
	var result cataloghttp.Package
	err := runtime.snapshot(ctx, access, cataloghttp.PermissionRead, "catalog.package.get",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			value, err := runtime.registry.GetPackage(scoped, scope, id.String())
			result = cataloghttp.RepresentPackage(value)
			return err
		})
	return result, err
}

func (runtime *HTTPRuntime) ListPackageVersions(ctx context.Context, access cataloghttp.Access,
	id uuid.UUID,
) ([]cataloghttp.PackageVersion, error) {
	var result []cataloghttp.PackageVersion
	err := runtime.snapshot(ctx, access, cataloghttp.PermissionRead, "catalog.package_version.list",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			values, err := runtime.registry.CompileCandidates(scoped, scope, id.String(), 64)
			if err != nil {
				return err
			}
			result = make([]cataloghttp.PackageVersion, 0, len(values))
			for _, value := range values {
				result = append(result, cataloghttp.RepresentVersion(value))
			}
			return nil
		})
	return result, err
}

func (runtime *HTTPRuntime) GetPackageVersion(ctx context.Context, access cataloghttp.Access,
	id uuid.UUID, version string,
) (cataloghttp.PackageVersion, error) {
	var result cataloghttp.PackageVersion
	err := runtime.snapshot(ctx, access, cataloghttp.PermissionRead, "catalog.package_version.get",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			value, err := runtime.registry.ResolveVersion(scoped, scope, id.String(), version)
			result = cataloghttp.RepresentVersion(value)
			return err
		})
	return result, err
}
