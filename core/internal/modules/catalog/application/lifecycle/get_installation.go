package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) GetInstallation(
	ctx context.Context,
	scope platformtx.TenantScope,
	packageID uuid.UUID,
) (catalogdomain.Installation, error) {
	if packageID == uuid.Nil {
		return catalogdomain.Installation{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, ScopeSnapshot, operationGetInstallation, ActionRead)
	if err != nil {
		return catalogdomain.Installation{}, err
	}
	identity, _ := packagecontract.ParsePackageID(packageID.String())
	value, err := store.GetInstallation(ctx, identity)
	if err != nil {
		return catalogdomain.Installation{}, ErrNotFound
	}
	if !value.ID.Valid() || value.PackageID.String() != packageID.String() {
		return catalogdomain.Installation{}, ErrUnavailable
	}
	return value, nil
}
