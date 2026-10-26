package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) Retry(
	ctx context.Context,
	scope platformtx.TenantScope,
	request RetryRequest,
) (OperationHandle, error) {
	if request.OperationID == uuid.Nil || !validIdempotency(request.IdempotencyKey) {
		return OperationHandle{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand, operationRetry, ActionInstall)
	if err != nil {
		return OperationHandle{}, err
	}
	previous, err := store.ReadOperation(ctx, request.OperationID)
	if err != nil {
		return OperationHandle{}, ErrNotFound
	}
	if previous.State != catalogdomain.OperationFailed && previous.State != catalogdomain.OperationBlocked {
		return OperationHandle{}, ErrConflict
	}
	installation, err := store.GetInstallation(ctx, previous.Plan.PackageID)
	if err != nil || installation.State != catalogdomain.InstallationFailed {
		return OperationHandle{}, ErrConflict
	}
	retrying, err := installation.Retry()
	if err != nil {
		return OperationHandle{}, ErrConflict
	}
	packageID, err := uuidFromCanonical(previous.Plan.PackageID.String())
	if err != nil {
		return OperationHandle{}, err
	}
	previousState := installation.State
	return application.start(ctx, scope, store, identity, startInput{
		Kind: catalogdomain.OperationRetry, PackageID: packageID,
		Version: previous.Plan.Version.String(), IdempotencyKey: request.IdempotencyKey,
		Installation: retrying, PreviousOperation: &request.OperationID,
		PreviousState: &previousState, MutateInstallation: true,
		RequestDocument: struct{ Operation, Previous string }{"retry", request.OperationID.String()},
	})
}
