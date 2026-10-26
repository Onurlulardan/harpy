package registry

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
)

//go:embed sql/platform/get_package_by_slug.sql
var getPackageBySlugSQL string

func (storage *store) ReadPackageBySlug(
	ctx context.Context, slug string,
) (packageregistry.Package, error) {
	if !slugPattern.MatchString(slug) {
		return packageregistry.Package{}, packageregistry.ErrInvalid
	}
	if !storage.usable(ctx) {
		return packageregistry.Package{}, packageregistry.ErrUnavailable
	}
	var value packageregistry.Package
	err := storage.queryRow(ctx, getPackageBySlugSQL, slug).Scan(&value.ID, &value.Slug, &value.CreatedAt)
	if err != nil {
		if storage.notFound(err) {
			return packageregistry.Package{}, packageregistry.ErrNotFound
		}
		return packageregistry.Package{}, packageregistry.ErrUnavailable
	}
	value.CreatedAt = value.CreatedAt.UTC()
	if !validPackageID(value.ID) || value.Slug != slug || value.CreatedAt.IsZero() {
		return packageregistry.Package{}, packageregistry.ErrUnavailable
	}
	return value, nil
}
