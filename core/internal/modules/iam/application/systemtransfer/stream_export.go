package systemtransfer

import (
	"context"
	"errors"
	"io"
)

func (worker *ExportWorker) streamExport(
	ctx context.Context,
	claim ExportJobClaim,
	attempt ExportAttempt,
) error {
	spool, err := worker.reserveExportPlaintext(ctx, attempt)
	if err != nil {
		return err
	}
	snapshotContext, cancelSnapshot := context.WithTimeout(ctx, MaximumExportSnapshotAge)
	progress, err := worker.writeExportPlaintext(
		snapshotContext, claim, attempt, spool,
	)
	cancelSnapshot()
	if err != nil {
		return errors.Join(err, discardExportPlaintext(spool))
	}
	reader, err := spool.OpenReader()
	if err != nil {
		return errors.Join(ErrExportUnavailable, err, discardExportPlaintext(spool))
	}
	receiptID, err := worker.newExportEffectReceiptID(attempt.AttemptID)
	if err != nil {
		return errors.Join(err, closeAndDiscardExportPlaintext(reader, spool))
	}
	callbacks := exportArtifactCallbacks{worker: worker, claim: claim,
		attempt: attempt, progress: progress, receiptID: receiptID,
		reader: reader, spool: spool}
	request := ExportArtifactSealRequest{TenantID: attempt.TenantID,
		ExportID: attempt.ExportID, EffectID: attempt.AttemptID,
		EffectReceiptID:   receiptID,
		AttemptGeneration: attempt.AttemptGeneration, Format: attempt.Format,
		Plaintext: reader, MaximumPlaintextBytes: exportPlaintextMaximumBytes()}
	if request.Validate() != nil {
		return errors.Join(ErrExportWorkerInvalid,
			closeAndDiscardExportPlaintext(reader, spool))
	}
	sealErr := worker.artifacts.SealExport(ctx, request,
		callbacks.begin, callbacks.commit)
	return callbacks.finish(ctx, sealErr)
}

func (worker *ExportWorker) reserveExportPlaintext(
	ctx context.Context,
	attempt ExportAttempt,
) (ExportPlaintextSpool, error) {
	identity := ExportPlaintextSpoolIdentity{TenantID: attempt.TenantID,
		ExportID: attempt.ExportID, AttemptGeneration: attempt.AttemptGeneration}
	spool, err := worker.spools.ReserveExportPlaintext(
		ctx, identity, exportPlaintextMaximumBytes(),
	)
	if err != nil || nilDependency(spool) {
		return nil, errors.Join(ErrExportCapacity, err)
	}
	return spool, nil
}

func (worker *ExportWorker) writeExportPlaintext(
	ctx context.Context,
	claim ExportJobClaim,
	attempt ExportAttempt,
	spool ExportPlaintextSpool,
) (ExportProgress, error) {
	writer, err := spool.OpenWriter()
	if err != nil {
		return ExportProgress{}, errors.Join(ErrExportUnavailable, err)
	}
	progress := ExportProgress{}
	writeErr := worker.writeExportSnapshot(ctx, claim, attempt, writer, &progress)
	closeErr := writer.Close()
	if writeErr != nil {
		return ExportProgress{}, errors.Join(writeErr, closeErr)
	}
	if closeErr != nil {
		return ExportProgress{}, errors.Join(ErrExportUnavailable, closeErr)
	}
	return progress, nil
}

func closeAndDiscardExportPlaintext(
	reader io.Closer,
	spool ExportPlaintextSpool,
) error {
	return errors.Join(reader.Close(), discardExportPlaintext(spool))
}

func discardExportPlaintext(spool ExportPlaintextSpool) error {
	if err := spool.Discard(); err != nil {
		return errors.Join(ErrExportUnavailable, err)
	}
	return nil
}

func (worker *ExportWorker) writeExportSnapshot(
	ctx context.Context,
	claim ExportJobClaim,
	attempt ExportAttempt,
	destination io.Writer,
	progress *ExportProgress,
) error {
	invoked := false
	err := worker.snapshots.WithPinnedSnapshot(ctx, attempt,
		func(scoped context.Context, snapshot ExportSnapshot) error {
			if invoked {
				return ErrExportCatalogPinStale
			}
			if validateErr := snapshot.Validate(attempt); validateErr != nil {
				return validateErr
			}
			invoked = true
			return worker.serializeExportSnapshot(scoped, claim, attempt,
				snapshot, destination, progress)
		})
	if err != nil {
		return err
	}
	if !invoked {
		return ErrExportStateConflict
	}
	return nil
}
