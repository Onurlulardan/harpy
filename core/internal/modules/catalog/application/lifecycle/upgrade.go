package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) Upgrade(
	ctx context.Context,
	scope platformtx.TenantScope,
	request UpgradeRequest,
) (OperationHandle, error) {
	if request.PackageID == uuid.Nil || request.Version == "" || !validIdempotency(request.IdempotencyKey) {
		return OperationHandle{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand, operationUpgrade, ActionUpgrade)
	if err != nil {
		return OperationHandle{}, err
	}
	packageID, _ := packagecontract.ParsePackageID(request.PackageID.String())
	installation, err := store.GetInstallation(ctx, packageID)
	if err != nil {
		return OperationHandle{}, ErrNotFound
	}
	target, err := packagecontract.ParseVersion(request.Version)
	if err != nil || target.Compare(installation.ActiveVersion) <= 0 ||
		(installation.State != catalogdomain.InstallationActive &&
			installation.State != catalogdomain.InstallationDisabled) {
		return OperationHandle{}, ErrConflict
	}
	return application.start(ctx, scope, store, identity, startInput{
		Kind: catalogdomain.OperationUpgrade, PackageID: request.PackageID,
		Version: request.Version, IdempotencyKey: request.IdempotencyKey,
		Installation: installation,
		RequestDocument: struct{ Operation, PackageID, Version string }{
			"upgrade", request.PackageID.String(), request.Version,
		},
	})
}
