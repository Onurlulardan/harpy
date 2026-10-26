package registry

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
)

//go:embed sql/platform/get_package.sql
var getPackageSQL string

func (storage *store) GetPackage(
	ctx context.Context, packageID string,
) (packageregistry.Package, error) {
	if !validPackageID(packageID) {
		return packageregistry.Package{}, packageregistry.ErrInvalid
	}
	if !storage.usable(ctx) {
		return packageregistry.Package{}, packageregistry.ErrUnavailable
	}
	var value packageregistry.Package
	err := storage.queryRow(ctx, getPackageSQL, packageID).Scan(&value.ID, &value.Slug, &value.CreatedAt)
	if err != nil {
		if storage.notFound(err) {
			return packageregistry.Package{}, packageregistry.ErrNotFound
		}
		return packageregistry.Package{}, packageregistry.ErrUnavailable
	}
	value.CreatedAt = value.CreatedAt.UTC()
	if value.ID != packageID || value.Slug == "" || value.CreatedAt.IsZero() {
		return packageregistry.Package{}, packageregistry.ErrUnavailable
	}
	return value, nil
}
