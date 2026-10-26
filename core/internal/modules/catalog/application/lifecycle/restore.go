package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) Restore(
	ctx context.Context,
	scope platformtx.TenantScope,
	command PackageCommand,
) (OperationHandle, error) {
	if command.PackageID == uuid.Nil || !validIdempotency(command.IdempotencyKey) {
		return OperationHandle{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand, operationRestore, ActionRestore)
	if err != nil {
		return OperationHandle{}, err
	}
	packageID, _ := packagecontract.ParsePackageID(command.PackageID.String())
	installation, err := store.GetInstallation(ctx, packageID)
	if err != nil {
		return OperationHandle{}, ErrNotFound
	}
	if installation.State != catalogdomain.InstallationRetired {
		return OperationHandle{}, ErrConflict
	}
	return application.startExisting(ctx, scope, store, identity, installation,
		catalogdomain.OperationRestore, command.IdempotencyKey, uuid.Nil)
}
