package systemtransfer

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
)

const exportOutcomePersistenceTimeout = 10 * time.Second

type exportArtifactCallbacks struct {
	worker        *ExportWorker
	claim         ExportJobClaim
	attempt       ExportAttempt
	progress      ExportProgress
	receiptID     uuid.UUID
	reader        io.ReadCloser
	spool         ExportPlaintextSpool
	effect        ExportArtifactEffectStart
	artifact      ExportArtifactMetadata
	began         bool
	started       bool
	committed     bool
	artifactKnown bool
	ready         bool
	cleaned       bool
}

func (callbacks *exportArtifactCallbacks) begin(
	ctx context.Context,
	evidence ExportArtifactBeginEvidence,
) error {
	if callbacks.began || evidence.Validate() != nil ||
		evidence.EffectID != callbacks.attempt.AttemptID {
		return ErrExportArtifactTampered
	}
	callbacks.began = true
	start := ExportArtifactEffectStart{Attempt: callbacks.attempt,
		EffectID: evidence.EffectID, EffectReceiptID: callbacks.receiptID,
		ObjectKey: evidence.ObjectKey, RequestSHA256: evidence.RequestSHA256,
		StartedAt: evidence.StartedAt, Draft: evidence.Draft}
	if start.Validate() != nil {
		return ErrExportArtifactTampered
	}
	if err := callbacks.worker.markArtifactEffectStarted(
		ctx, callbacks.claim, start,
	); err != nil {
		return err
	}
	callbacks.effect = start
	callbacks.started = true
	return nil
}

func (callbacks *exportArtifactCallbacks) commit(
	ctx context.Context,
	metadata ExportArtifactMetadata,
) error {
	if !callbacks.started || callbacks.committed ||
		exportArtifactMatches(callbacks.effect, metadata) != nil {
		return ErrExportArtifactTampered
	}
	callbacks.committed = true
	callbacks.artifact = metadata
	callbacks.artifactKnown = true
	callbacks.cleaned = true
	if err := closeAndDiscardExportPlaintext(
		callbacks.reader, callbacks.spool,
	); err != nil {
		return err
	}
	err := callbacks.worker.commitExport(ctx, callbacks.claim,
		callbacks.attempt, callbacks.progress, callbacks.effect, metadata)
	callbacks.ready = err == nil
	return err
}

func (callbacks *exportArtifactCallbacks) finish(
	ctx context.Context,
	sealErr error,
) error {
	if callbacks.ready {
		return nil
	}
	if !callbacks.cleaned {
		callbacks.cleaned = true
		sealErr = errors.Join(sealErr,
			closeAndDiscardExportPlaintext(callbacks.reader, callbacks.spool))
	}
	if !callbacks.started {
		if sealErr == nil {
			return ErrExportStateConflict
		}
		return sealErr
	}
	if sealErr == nil {
		sealErr = ErrExportStateConflict
	}
	finalizeContext, cancel := context.WithTimeout(
		context.WithoutCancel(ctx), exportOutcomePersistenceTimeout,
	)
	defer cancel()
	markErr := callbacks.worker.markExportOutcomeUnknown(
		finalizeContext, callbacks.effect, callbacks.artifact,
		callbacks.artifactKnown,
	)
	if markErr != nil {
		return errors.Join(ErrExportOutcomeUnknown,
			ErrExportOutcomeUnknownPersist, sealErr, markErr)
	}
	return ErrExportOutcomeUnknown
}

func exportArtifactMatches(
	start ExportArtifactEffectStart,
	metadata ExportArtifactMetadata,
) error {
	if metadata.Validate() != nil || metadata.EffectID != start.EffectID ||
		metadata.EffectReceiptID != start.EffectReceiptID ||
		metadata.ObjectKey != start.ObjectKey ||
		metadata.RequestSHA256 != start.RequestSHA256 ||
		!exportDraftMatchesMetadata(start.Draft, metadata) ||
		metadata.AttemptGeneration != start.Attempt.AttemptGeneration {
		return ErrExportArtifactTampered
	}
	return nil
}
