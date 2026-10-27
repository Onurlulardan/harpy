package lifecycle

import (
	"context"
	_ "embed"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

//go:embed sql/tenant/list_dependency_blockers.sql
var listDependencyBlockersSQL string

func (storage *store) ResolveDependencyBlockers(
	ctx context.Context, packageID packagecontract.PackageID,
) ([]catalogapp.DependencyBlocker, error) {
	if !packageID.Valid() {
		return nil, catalogapp.ErrInvalid
	}
	if !storage.usable(ctx, false) {
		return nil, unavailable()
	}
	resultRows, err := storage.query(ctx, listDependencyBlockersSQL, packageID.String())
	if err != nil {
		return nil, unavailable()
	}
	defer resultRows.Close()
	result := make([]catalogapp.DependencyBlocker, 0)
	for resultRows.Next() {
		var packageRaw, versionRaw string
		var required bool
		if err = resultRows.Scan(&packageRaw, &versionRaw, &required); err != nil {
			return nil, unavailable()
		}
		dependentID, packageErr := packagecontract.ParsePackageID(packageRaw)
		version, versionErr := packagecontract.ParseVersion(versionRaw)
		if packageErr != nil || versionErr != nil || dependentID.String() == packageID.String() {
			return nil, unavailable()
		}
		result = append(result, catalogapp.DependencyBlocker{
			PackageID: dependentID, Version: version, Required: required,
		})
	}
	if resultRows.Err() != nil {
		return nil, unavailable()
	}
	return result, nil
}
