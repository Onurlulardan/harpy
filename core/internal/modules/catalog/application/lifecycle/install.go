package lifecycle

import (
	"context"
	"errors"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) Install(
	ctx context.Context,
	scope platformtx.TenantScope,
	request InstallRequest,
) (OperationHandle, error) {
	if request.PackageID == uuid.Nil || request.Version == "" || !validIdempotency(request.IdempotencyKey) {
		return OperationHandle{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand, operationInstall, ActionInstall)
	if err != nil {
		return OperationHandle{}, err
	}
	packageID, err := packagecontract.ParsePackageID(request.PackageID.String())
	if err != nil {
		return OperationHandle{}, ErrInvalid
	}
	if _, err = store.GetInstallation(ctx, packageID); err == nil || !errors.Is(err, ErrNotFound) {
		return OperationHandle{}, ErrConflict
	}
	installationID, err := application.nextIdentity()
	if err != nil {
		return OperationHandle{}, err
	}
	stableInstallationID, err := stableIdentity(installationID)
	if err != nil {
		return OperationHandle{}, err
	}
	installation, err := catalogdomain.NewInstalling(stableInstallationID, packageID)
	if err != nil {
		return OperationHandle{}, ErrInvalid
	}
	return application.start(ctx, scope, store, identity, startInput{
		Kind: catalogdomain.OperationInstall, PackageID: request.PackageID,
		Version: request.Version, IdempotencyKey: request.IdempotencyKey,
		Installation: installation, MutateInstallation: true,
		RequestDocument: struct{ Operation, PackageID, Version string }{
			"install", request.PackageID.String(), request.Version,
		},
	})
}
