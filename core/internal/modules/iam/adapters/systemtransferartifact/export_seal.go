package systemtransferartifact

import (
	"context"
	"errors"
	"time"

	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (runtime *Runtime) SealExport(
	ctx context.Context,
	request systemtransfer.ExportArtifactSealRequest,
	begin systemtransfer.ExportArtifactBegin,
	commit systemtransfer.ExportArtifactCommit,
) (err error) {
	if runtime == nil || request.Validate() != nil || begin == nil || commit == nil ||
		!runtimeInputContext(ctx) {
		return errInvalid
	}
	draft, err := runtime.sealLocal(ctx, localSealRequest{
		tenantID: request.TenantID, artifactID: request.ExportID,
		leaseOwner: request.EffectID, generation: request.AttemptGeneration,
		format: request.Format, prefix: exportPrefix, plaintext: request.Plaintext,
		maximumPlaintextBytes: request.MaximumPlaintextBytes,
		stableReservation:     true, keys: runtime.exportKeys,
	})
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, draft.lease.Discard()) }()
	requestSHA := exportUploadRequestSHA(draft.objectKey,
		draft.ciphertext.CiphertextSHA256, int64(draft.ciphertext.CiphertextBytes))
	startedAt, err := runtime.evidenceNow()
	if err != nil {
		return err
	}
	if err = begin(ctx, systemtransfer.ExportArtifactBeginEvidence{
		EffectID: request.EffectID, ObjectKey: draft.objectKey,
		RequestSHA256: requestSHA, StartedAt: startedAt,
		Draft: systemtransfer.ExportArtifactDraftEvidence{
			KeyID:            draft.envelope.KEKID,
			HeaderSHA256:     draft.ciphertext.HeaderSHA256,
			CiphertextSHA256: draft.ciphertext.CiphertextSHA256,
			CiphertextBytes:  int64(draft.ciphertext.CiphertextBytes),
			WrappedDEK:       append([]byte(nil), draft.envelope.Ciphertext...),
			WrapNonce:        draft.envelope.Nonce,
		},
	}); err != nil {
		return err
	}
	upload, err := publishDraft(ctx, runtime.exports, draft)
	if err != nil {
		return err
	}
	metadata := systemtransfer.ExportArtifactMetadata{
		ObjectKey: upload.Identity.Key, ObjectVersion: upload.Identity.Version,
		KeyID: draft.envelope.KEKID, ProviderRequestID: upload.ProviderRequestID,
		EffectID: request.EffectID, EffectReceiptID: request.EffectReceiptID,
		HeaderSHA256:     draft.ciphertext.HeaderSHA256,
		CiphertextSHA256: draft.ciphertext.CiphertextSHA256,
		RequestSHA256:    requestSHA,
		ResultSHA256: exportUploadResultSHA(upload.Identity.Key,
			upload.Identity.Version, upload.ProviderRequestID,
			int64(draft.ciphertext.CiphertextBytes)),
		CiphertextBytes:   int64(draft.ciphertext.CiphertextBytes),
		WrappedDEK:        append([]byte(nil), draft.envelope.Ciphertext...),
		WrapNonce:         draft.envelope.Nonce,
		AttemptGeneration: request.AttemptGeneration,
		SealedAt:          upload.UploadedAt.UTC().Truncate(time.Microsecond),
	}
	if metadata.Validate() != nil {
		return errOutcomeUnknown
	}
	return commit(ctx, metadata)
}

func (runtime *Runtime) evidenceNow() (time.Time, error) {
	if runtime == nil || nilDependency(runtime.clock) {
		return time.Time{}, errUnavailable
	}
	value := runtime.clock.Now().UTC().Truncate(time.Microsecond)
	if !validEvidenceTime(value) {
		return time.Time{}, errUnavailable
	}
	return value, nil
}

var _ systemtransfer.ExportArtifactSealer = (*Runtime)(nil)
