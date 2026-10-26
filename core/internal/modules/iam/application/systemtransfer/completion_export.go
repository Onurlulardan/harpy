package systemtransfer

import (
	"context"
	"errors"
)

func (worker *ExportWorker) commitExport(
	ctx context.Context,
	claim ExportJobClaim,
	attempt ExportAttempt,
	progress ExportProgress,
	effect ExportArtifactEffectStart,
	metadata ExportArtifactMetadata,
) error {
	if effect.Validate() != nil || effect.Attempt.AttemptID != attempt.AttemptID ||
		exportArtifactMatches(effect, metadata) != nil || progress.Validate() != nil {
		return ErrExportArtifactTampered
	}
	if err := worker.checkLease(ctx, claim); err != nil {
		return err
	}
	return worker.scopes.InFreshExportCheckpoint(ctx, attempt,
		func(scoped context.Context, scope ExportScope) error {
			store, err := worker.authorizedExportStore(scoped, scope, attempt)
			if err != nil {
				return err
			}
			replay, err := store.PersistOutboundSuccess(scoped, ExportCompletion{
				Attempt: attempt, Progress: progress, Artifact: metadata})
			if err != nil {
				return err
			}
			if replay && metadata.AttemptGeneration != attempt.AttemptGeneration {
				return ErrExportStateConflict
			}
			return nil
		})
}

func normalizeExportCommitError(err error) error {
	if err == nil {
		return nil
	}
	for _, sentinel := range []error{ErrExportAuthorizationRevoked,
		ErrExportCancelled, ErrExportLeaseLost, ErrExportLimit,
		ErrExportQueryInvalid, ErrExportCatalogPinStale,
		ErrExportArtifactTampered, ErrExportStateConflict} {
		if errors.Is(err, sentinel) {
			return err
		}
	}
	return errors.Join(ErrExportUnavailable, err)
}
