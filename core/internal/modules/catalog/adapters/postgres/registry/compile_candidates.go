package registry

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
)

//go:embed sql/platform/list_versions.sql
var listVersionsSQL string

func (storage *store) CompileCandidates(
	ctx context.Context, packageID string, limit int,
) ([]packageregistry.SealedVersion, error) {
	if !validPackageID(packageID) || limit < 1 || limit > 64 {
		return nil, packageregistry.ErrInvalid
	}
	if !storage.usable(ctx) {
		return nil, packageregistry.ErrUnavailable
	}
	resultRows, err := storage.query(ctx, listVersionsSQL, packageID, limit)
	if err != nil {
		return nil, packageregistry.ErrUnavailable
	}
	defer resultRows.Close()
	result := make([]packageregistry.SealedVersion, 0, limit)
	for resultRows.Next() {
		value, scanErr := scanVersion(resultRows)
		if scanErr != nil || !validVersion(value) || value.PackageID != packageID {
			return nil, packageregistry.ErrUnavailable
		}
		result = append(result, value)
	}
	if resultRows.Err() != nil {
		return nil, packageregistry.ErrUnavailable
	}
	return result, nil
}
