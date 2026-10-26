package systemtransfer

import (
	"context"
	"errors"
)

type ExportReconciliationProcessor struct {
	store      ExportReconciliationStore
	reconciler ExportArtifactOutcomeReconciler
	clock      ExportClock
	backoff    ExportReconciliationBackoff
}

func NewSystemExportReconciliationProcessor(
	store ExportReconciliationStore,
	reconciler ExportArtifactOutcomeReconciler,
	clock ExportClock,
	backoff ExportReconciliationBackoff,
) (*ExportReconciliationProcessor, error) {
	if nilDependency(store) || nilDependency(reconciler) ||
		nilDependency(clock) || nilDependency(backoff) {
		return nil, ErrExportReconciliationInvalid
	}
	return &ExportReconciliationProcessor{store: store, reconciler: reconciler,
		clock: clock, backoff: backoff}, nil
}

func (processor *ExportReconciliationProcessor) ProcessExportReconciliation(
	ctx context.Context,
	claim ExportReconciliationClaim,
) error {
	if processor == nil || ctx == nil || ctx.Err() != nil || claim.Validate() != nil ||
		nilDependency(processor.store) || nilDependency(processor.reconciler) ||
		nilDependency(processor.clock) || nilDependency(processor.backoff) {
		return ErrExportReconciliationInvalid
	}
	observation, claimed, err := processor.store.AcquireReconciliation(ctx, claim)
	if err != nil {
		return errors.Join(ErrExportReconciliationPersist, err)
	}
	if !claimed {
		return nil
	}
	if observation.Validate(claim) != nil {
		return ErrExportReconciliationInvalid
	}
	result, reconcileErr := invokeExportArtifactReconciliation(
		ctx, processor.reconciler, observation.ExportEffect,
	)
	if reconcileErr == nil && result.Validate(observation.ExportEffect) == nil {
		switch result.State {
		case ExportArtifactSealed:
			return processor.resolveSealed(ctx, claim, observation, result)
		case ExportArtifactMissing:
			return processor.resolveMissing(ctx, claim, observation, result)
		}
	}
	return processor.resolveUnknown(ctx, claim, observation)
}

func invokeExportArtifactReconciliation(
	ctx context.Context,
	reconciler ExportArtifactOutcomeReconciler,
	effect ExportArtifactEffectStart,
) (result ExportArtifactReconciliation, err error) {
	defer func() {
		if recover() != nil {
			result = ExportArtifactReconciliation{}
			err = ErrExportOutcomeUnknown
		}
	}()
	return reconciler.ReconcileExportArtifact(ctx, effect)
}
