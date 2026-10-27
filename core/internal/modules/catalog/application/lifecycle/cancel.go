package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) Cancel(
	ctx context.Context,
	scope platformtx.TenantScope,
	request CancelRequest,
) (catalogdomain.Operation, error) {
	if request.OperationID == uuid.Nil || !validIdempotency(request.IdempotencyKey) {
		return catalogdomain.Operation{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand, operationCancel, ActionInstall)
	if err != nil {
		return catalogdomain.Operation{}, err
	}
	current, err := store.ReadOperation(ctx, request.OperationID)
	if err != nil {
		return catalogdomain.Operation{}, ErrNotFound
	}
	next, err := current.RequestCancel()
	if err != nil {
		return catalogdomain.Operation{}, ErrConflict
	}
	stored, err := ChangeOperation(ctx, store, OperationMutation{
		Operation: next, ExpectedRevision: current.Revision,
	})
	if err != nil || stored.Revision != next.Revision {
		return catalogdomain.Operation{}, ErrUnavailable
	}
	if err = application.jobs.Abort(ctx, scope, request.OperationID); err != nil {
		return catalogdomain.Operation{}, err
	}
	if stored.State == catalogdomain.OperationCancelled {
		if err = application.cancelInstallation(ctx, store, identity, request, stored); err != nil {
			return catalogdomain.Operation{}, err
		}
	}
	return stored, nil
}

func (application *Application) cancelInstallation(
	ctx context.Context,
	store Store,
	identity ScopeIdentity,
	request CancelRequest,
	operation catalogdomain.Operation,
) error {
	installation, err := store.GetInstallation(ctx, operation.Plan.PackageID)
	if err != nil {
		return err
	}
	currentRevision := installation.RowVersion
	switch {
	case operation.Kind == catalogdomain.OperationPurge &&
		installation.State == catalogdomain.InstallationPurgePending:
		installation, err = installation.CancelPurge()
	case (operation.Kind == catalogdomain.OperationInstall || operation.Kind == catalogdomain.OperationRetry) &&
		installation.State == catalogdomain.InstallationInstalling:
		installation, err = installation.Fail()
	default:
		return nil
	}
	if err != nil {
		return ErrConflict
	}
	now, timeErr := application.now()
	if timeErr != nil {
		return timeErr
	}
	digest, digestErr := requestDigest(struct{ OperationID string }{request.OperationID.String()})
	if digestErr != nil {
		return digestErr
	}
	_, err = UpdateInstallation(ctx, store, InstallationMutation{
		Installation: installation, ExpectedRevision: currentRevision,
		ActorID: identity.ActorID, IdempotencyKey: request.IdempotencyKey,
		RequestSHA256: digest, OccurredAt: now,
	})
	return err
}
