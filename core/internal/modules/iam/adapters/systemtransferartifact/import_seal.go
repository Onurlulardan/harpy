package systemtransferartifact

import (
	"context"
	"errors"
	"time"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

const importCleanupTimeout = 10 * time.Second

func (runtime *Runtime) SealImportSource(
	ctx context.Context,
	request systemtransfer.SourceSealRequest,
	register systemtransfer.SourceRegistration,
) (record systemtransfer.SourceRecord, err error) {
	if runtime == nil || !validImportSeal(ctx, request, register) {
		return record, errInvalid
	}
	draft, err := runtime.sealLocal(ctx, localSealRequest{
		tenantID: request.TenantID, artifactID: request.SourceID,
		leaseOwner: request.ActorID, generation: 1, format: request.Format,
		prefix: importPrefix, plaintext: request.Body,
		maximumPlaintextBytes: request.MaximumBytes,
		stableReservation:     true, analyze: true, keys: runtime.importKeys,
	})
	if err != nil {
		return record, err
	}
	defer func() { err = errors.Join(err, draft.lease.Discard()) }()
	upload, err := publishDraft(ctx, runtime.imports, draft)
	if err != nil {
		return record, err
	}
	artifact := sourceArtifact(draft, upload)
	record, registerErr := register(ctx, artifact)
	if registerErr != nil {
		if !registrationDefinitelyRejected(registerErr) {
			return record, errors.Join(errOutcomeUnknown, registerErr)
		}
		if cleanupErr := runtime.deleteImportExact(ctx, upload.Identity); cleanupErr != nil {
			return record, errors.Join(errOutcomeUnknown, registerErr, cleanupErr)
		}
		return record, registerErr
	}
	if !sourceRecordMatches(record, request, draft) {
		if record.SourceID != request.SourceID {
			return systemtransfer.SourceRecord{}, errors.Join(errOutcomeUnknown,
				runtime.deleteImportExact(ctx, upload.Identity))
		}
		return systemtransfer.SourceRecord{}, errOutcomeUnknown
	}
	if record.SourceID != request.SourceID {
		if cleanupErr := runtime.deleteImportExact(ctx, upload.Identity); cleanupErr != nil {
			return systemtransfer.SourceRecord{}, errors.Join(errOutcomeUnknown, cleanupErr)
		}
	}
	record.Analysis = draft.analysis.OwnedCopy()
	return record, nil
}

func sourceArtifact(draft sealedDraft,
	upload artifactobject.UploadResult) systemtransfer.SourceArtifact {
	return systemtransfer.SourceArtifact{
		ObjectKey: upload.Identity.Key, ObjectVersion: upload.Identity.Version,
		KeyID: draft.envelope.KEKID, ProviderRequestID: upload.ProviderRequestID,
		HeaderSHA256:     draft.ciphertext.HeaderSHA256,
		CiphertextSHA256: draft.ciphertext.CiphertextSHA256,
		PlaintextSHA256:  draft.plaintextSHA256,
		CiphertextBytes:  int64(draft.ciphertext.CiphertextBytes),
		PlaintextBytes:   draft.plaintextBytes,
		WrappedDEK:       append([]byte(nil), draft.envelope.Ciphertext...),
		WrapNonce:        draft.envelope.Nonce, UploadedAt: upload.UploadedAt.UTC(),
	}
}

func (runtime *Runtime) deleteImportExact(
	ctx context.Context,
	identity artifactobject.Identity,
) error {
	cleanupContext, cancel := context.WithTimeout(
		context.WithoutCancel(ctx), importCleanupTimeout,
	)
	defer cancel()
	result, err := runtime.imports.DeleteExact(cleanupContext, identity)
	if err != nil {
		return err
	}
	if result.ProviderRequestID == "" || result.DeleteResponseSHA256 == [32]byte{} ||
		result.AbsenceProbeSHA256 == [32]byte{} || !validEvidenceTime(result.VerifiedAbsentAt) {
		return errOutcomeUnknown
	}
	return nil
}

func validImportSeal(ctx context.Context, request systemtransfer.SourceSealRequest,
	register systemtransfer.SourceRegistration) bool {
	return runtimeInputContext(ctx) && request.TenantID != [16]byte{} &&
		request.ActorID != [16]byte{} && request.SourceID.Version() == 7 &&
		request.Resource.Valid() && (request.Format == "csv" || request.Format == "jsonl") &&
		request.IdempotencySHA256 != [32]byte{} && validEvidenceTime(request.UploadedAt) &&
		request.Body != nil && request.MaximumBytes > 0 && request.MaximumBytes <= 536870912 &&
		register != nil
}

func validEvidenceTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Year() >= 2000 &&
		value.Nanosecond()%int(time.Microsecond) == 0
}
