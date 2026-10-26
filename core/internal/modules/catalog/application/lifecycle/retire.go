package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) Retire(
	ctx context.Context,
	scope platformtx.TenantScope,
	command PackageCommand,
) (OperationHandle, error) {
	if command.PackageID == uuid.Nil || !validIdempotency(command.IdempotencyKey) {
		return OperationHandle{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand, operationRetire, ActionRetire)
	if err != nil {
		return OperationHandle{}, err
	}
	packageID, _ := packagecontract.ParsePackageID(command.PackageID.String())
	installation, err := store.GetInstallation(ctx, packageID)
	if err != nil {
		return OperationHandle{}, ErrNotFound
	}
	if installation.State != catalogdomain.InstallationActive &&
		installation.State != catalogdomain.InstallationDisabled {
		return OperationHandle{}, ErrConflict
	}
	blockers, err := ResolveDependencyBlockers(ctx, store, packageID)
	if err != nil {
		return OperationHandle{}, err
	}
	for _, blocker := range blockers {
		if blocker.Required {
			return OperationHandle{}, ErrDependencyBlocked
		}
	}
	return application.startExisting(ctx, scope, store, identity, installation,
		catalogdomain.OperationRetire, command.IdempotencyKey, uuid.Nil)
}
