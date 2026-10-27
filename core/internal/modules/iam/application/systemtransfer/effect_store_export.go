package systemtransfer

import (
	"context"
	"errors"
)

func (worker *ExportWorker) markArtifactEffectStarted(
	ctx context.Context,
	claim ExportJobClaim,
	start ExportArtifactEffectStart,
) error {
	if start.Validate() != nil {
		return ErrExportArtifactTampered
	}
	if err := worker.checkLease(ctx, claim); err != nil {
		return err
	}
	return worker.scopes.InFreshExportCheckpoint(ctx, start.Attempt,
		func(scoped context.Context, scope ExportScope) error {
			store, err := worker.authorizedExportStore(scoped, scope, start.Attempt)
			if err != nil {
				return err
			}
			_, err = store.BeginArtifactEffect(scoped, start)
			return err
		})
}

func (worker *ExportWorker) markExportOutcomeUnknown(
	ctx context.Context,
	effect ExportArtifactEffectStart,
	artifact ExportArtifactMetadata,
	artifactKnown bool,
) error {
	now, err := worker.exportNow()
	if err != nil {
		return err
	}
	record := ExportOutcomeUnknownRecord{Effect: effect, Artifact: artifact,
		ArtifactKnown: artifactKnown, At: now, Code: ExportOutcomeUnknownCode}
	if record.Validate() != nil {
		return ErrExportWorkerInvalid
	}
	return worker.scopes.InFreshExportCheckpoint(ctx, effect.Attempt,
		func(scoped context.Context, scope ExportScope) error {
			if scope.TenantID() != effect.Attempt.TenantID ||
				scope.ActorID() != effect.Attempt.ActorID {
				return ErrExportStateConflict
			}
			store, bindErr := worker.stores.BindExportWorker(scoped, scope.IAMScope())
			if bindErr != nil || nilDependency(store) {
				return errors.Join(ErrExportUnavailable, bindErr)
			}
			return store.PersistOutcomeUnknown(scoped, record)
		})
}
