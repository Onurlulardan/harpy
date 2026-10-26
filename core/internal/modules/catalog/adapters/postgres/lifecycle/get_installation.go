package lifecycle

import (
	"context"
	_ "embed"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

//go:embed sql/tenant/get_installation.sql
var getInstallationSQL string

func (storage *store) GetInstallation(
	ctx context.Context, packageID packagecontract.PackageID,
) (catalogdomain.Installation, error) {
	if !packageID.Valid() {
		return catalogdomain.Installation{}, catalogapp.ErrInvalid
	}
	if !storage.usable(ctx, false) {
		return catalogdomain.Installation{}, unavailable()
	}
	value, err := scanInstallation(storage.queryRow(ctx, getInstallationSQL, packageID.String()))
	if err != nil {
		if storage.notFound(err) {
			return catalogdomain.Installation{}, catalogapp.ErrNotFound
		}
		return catalogdomain.Installation{}, unavailable()
	}
	if value.PackageID.String() != packageID.String() {
		return catalogdomain.Installation{}, unavailable()
	}
	return value, nil
}
