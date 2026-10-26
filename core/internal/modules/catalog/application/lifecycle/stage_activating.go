package lifecycle

import (
	"context"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) stageActivating(
	ctx context.Context,
	scope platformtx.TenantScope,
	store Store,
	operation catalogdomain.Operation,
) (ResumeResult, error) {
	if demoOperation(operation.Kind) {
		return application.stageDemoActivating(ctx, store, operation)
	}
	installation, err := store.GetInstallation(ctx, operation.Plan.PackageID)
	if err != nil {
		return ResumeResult{}, ErrNotFound
	}
	currentRevision := installation.RowVersion
	action := operationOwnerAction(operation.Kind)
	if action != "purge" {
		if _, err = application.applyOwnerSteps(
			ctx, scope, store, operation, action, installation.ReferenceEpoch, false, 0,
		); err != nil {
			return ResumeResult{}, err
		}
	}
	switch operation.Kind {
	case catalogdomain.OperationInstall, catalogdomain.OperationRetry:
		installation, err = installation.Activate(operation.Plan.Version)
	case catalogdomain.OperationUpgrade:
		installation, err = installation.Upgrade(operation.Plan.Version)
	case catalogdomain.OperationRestore:
		installation, err = installation.Restore()
	case catalogdomain.OperationRetire:
		installation, err = installation.Retire()
	case catalogdomain.OperationPurge:
		installation, err = installation.FinishPurge()
	default:
		err = ErrInvalid
	}
	if err != nil {
		return ResumeResult{}, ErrConflict
	}
	now, err := application.now()
	if err != nil {
		return ResumeResult{}, err
	}
	if _, err = UpdateInstallation(ctx, store, InstallationMutation{
		Installation: installation, ExpectedRevision: currentRevision,
		OperationID: operation.ID,
		ActorID:     scope.ActorID(), IdempotencyKey: operation.ID.String() + ":activation",
		RequestSHA256: operation.Plan.SHA256, OccurredAt: now,
	}); err != nil {
		return ResumeResult{}, err
	}
	result, err := application.advanceOperation(ctx, store, operation, catalogdomain.OperationSucceeded)
	if err != nil {
		return ResumeResult{}, err
	}
	if err = application.events.Emit(ctx, scope, InstallationEvent{
		Type: eventType(operation.Kind), Version: "1", IdempotencyKey: operation.ID.String(),
		PlanSHA256: operation.Plan.SHA256, Installation: installation,
		OperationID: operation.ID, OccurredAt: now,
	}); err != nil {
		return ResumeResult{}, err
	}
	if operation.Kind == catalogdomain.OperationInstall || operation.Kind == catalogdomain.OperationUpgrade ||
		operation.Kind == catalogdomain.OperationRetry {
		if err = application.events.Emit(ctx, scope, InstallationEvent{
			Type: "catalog.menu.published", Version: "1",
			IdempotencyKey: operation.ID.String() + ":menu", PlanSHA256: operation.Plan.SHA256,
			Installation: installation,
			OperationID:  operation.ID, OccurredAt: now,
		}); err != nil {
			return ResumeResult{}, err
		}
	}
	return result, nil
}

func operationOwnerAction(kind catalogdomain.OperationKind) string {
	switch kind {
	case catalogdomain.OperationInstall, catalogdomain.OperationUpgrade, catalogdomain.OperationRetry:
		return "activate"
	case catalogdomain.OperationRestore:
		return "restore"
	case catalogdomain.OperationRetire:
		return "retire"
	case catalogdomain.OperationPurge:
		return "purge"
	default:
		return ""
	}
}

func eventType(kind catalogdomain.OperationKind) string {
	switch kind {
	case catalogdomain.OperationRetire:
		return "catalog.installation.retired"
	case catalogdomain.OperationPurge:
		return "catalog.installation.purged"
	default:
		return "catalog.installation.activated"
	}
}
