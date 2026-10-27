package systemtransfer

import (
	"context"
	"errors"
)

func (processor *ExportReconciliationProcessor) resolveSealed(
	ctx context.Context,
	claim ExportReconciliationClaim,
	observation ExportReconciliationObservation,
	result ExportArtifactReconciliation,
) error {
	receipt, err := newExportReconciliationReceipt(observation, result)
	if err != nil {
		return processor.resolveUnknown(ctx, claim, observation)
	}
	finalize, cancel := exportReconciliationFinalizeContext(ctx)
	defer cancel()
	if err = processor.store.PersistReconciliationSealed(
		finalize, claim, observation, result, receipt,
	); err != nil {
		return errors.Join(ErrExportReconciliationPersist, err)
	}
	return nil
}

func (processor *ExportReconciliationProcessor) resolveMissing(
	ctx context.Context,
	claim ExportReconciliationClaim,
	observation ExportReconciliationObservation,
	result ExportArtifactReconciliation,
) error {
	receipt, err := newExportReconciliationReceipt(observation, result)
	if err != nil {
		return processor.resolveUnknown(ctx, claim, observation)
	}
	finalize, cancel := exportReconciliationFinalizeContext(ctx)
	defer cancel()
	if err = processor.store.PersistReconciliationMissing(
		finalize, claim, observation, result, receipt,
		ExportReconciliationMissingCode,
	); err != nil {
		return errors.Join(ErrExportReconciliationPersist, err)
	}
	return nil
}

func exportReconciliationFinalizeContext(
	ctx context.Context,
) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx),
		exportOutcomePersistenceTimeout)
}
