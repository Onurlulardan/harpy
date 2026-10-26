package registry

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
)

//go:embed sql/platform/list_packages.sql
var listPackagesSQL string

func (storage *store) ListPackages(
	ctx context.Context, limit int, after string,
) ([]packageregistry.Package, error) {
	if limit < 1 || limit > 100 || after != "" && !validPackageID(after) {
		return nil, packageregistry.ErrInvalid
	}
	if !storage.usable(ctx) {
		return nil, packageregistry.ErrUnavailable
	}
	var cursor any
	if after != "" {
		cursor = after
	}
	resultRows, err := storage.query(ctx, listPackagesSQL, cursor, limit)
	if err != nil {
		return nil, packageregistry.ErrUnavailable
	}
	defer resultRows.Close()
	result := make([]packageregistry.Package, 0, limit)
	for resultRows.Next() {
		var value packageregistry.Package
		if err = resultRows.Scan(&value.ID, &value.Slug, &value.CreatedAt); err != nil {
			return nil, packageregistry.ErrUnavailable
		}
		value.CreatedAt = value.CreatedAt.UTC()
		if !validPackageID(value.ID) || !slugPattern.MatchString(value.Slug) || value.CreatedAt.IsZero() {
			return nil, packageregistry.ErrUnavailable
		}
		result = append(result, value)
	}
	if resultRows.Err() != nil {
		return nil, packageregistry.ErrUnavailable
	}
	return result, nil
}
