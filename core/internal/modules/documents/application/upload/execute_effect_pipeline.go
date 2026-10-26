package upload

import (
	"context"
	"io"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type EffectPipeline struct {
	keyring Keyring
	cipher  Cipher
	scanner Scanner
	objects ObjectStore
	clock   foundation.Clock
}

func NewEffectPipeline(keyring Keyring, cipher Cipher, scanner Scanner,
	objects ObjectStore, clock foundation.Clock,
) (*EffectPipeline, error) {
	for _, dependency := range []any{keyring, cipher, scanner, objects, clock} {
		if foundation.NilPort(dependency) {
			return nil, ErrUnavailable
		}
	}
	return &EffectPipeline{keyring: keyring, cipher: cipher, scanner: scanner,
		objects: objects, clock: clock}, nil
}

func (pipeline *EffectPipeline) Execute(ctx context.Context, plan AttemptPlan,
	content io.Reader,
) (EffectReceipt, *EffectFailure) {
	if pipeline == nil || ctx == nil || ctx.Err() != nil || content == nil ||
		!platformtx.NetworkAllowed(ctx) || !validAttemptPlan(plan) {
		return EffectReceipt{}, pipeline.failure(plan, FailureInvalid,
			"DOCUMENT_EFFECT_INPUT_INVALID", nil)
	}
	key, err := pipeline.keyring.CreateDataKey(ctx, plan.TenantID,
		plan.DocumentID, plan.CandidateVersionID)
	if err != nil || key.KeyID == "" || len(key.Wrapped) < 115 {
		clear(key.Plaintext[:])
		return EffectReceipt{}, pipeline.failure(plan, FailureTransient,
			"DOCUMENT_KEYRING_UNAVAILABLE", nil)
	}
	defer clear(key.Plaintext[:])
	scan, err := pipeline.scanner.Open(ctx, plan.OriginalName,
		plan.DeclaredMIME, plan.ExpectedBytes)
	if err != nil || scan == nil {
		return EffectReceipt{}, pipeline.failure(plan, FailureTransient,
			"DOCUMENT_SCAN_UNAVAILABLE", nil)
	}
	cipherReceipt, objectResult, effectErr := pipeline.stream(ctx, plan, content, scan, key)
	if effectErr != nil {
		return EffectReceipt{}, pipeline.failure(plan, FailureTransient,
			"DOCUMENT_OBJECT_WRITE_FAILED", nil)
	}
	object := domain.ObjectReceipt{Key: plan.ObjectKey, Version: objectResult.Version,
		KeyID: key.KeyID, Generation: objectResult.Generation,
		PlaintextBytes: plan.ExpectedBytes, CiphertextBytes: cipherReceipt.CiphertextBytes,
		PlaintextSHA256:  cipherReceipt.PlaintextSHA256,
		CiphertextSHA256: cipherReceipt.CiphertextSHA256, WrappedDEK: key.Wrapped}
	scanResult, scanErr := scan.Finish(ctx)
	if scanErr != nil {
		return EffectReceipt{}, pipeline.failure(plan, FailureTransient,
			"DOCUMENT_SCAN_UNAVAILABLE", &object)
	}
	if !scanResult.Clean {
		return EffectReceipt{}, pipeline.failure(plan, FailureRejected,
			"DOCUMENT_CONTENT_REJECTED", &object)
	}
	if cipherReceipt.PlaintextSHA256 != plan.ExpectedContentSHA256 {
		return EffectReceipt{}, pipeline.failure(plan, FailureInvalid,
			"DOCUMENT_CONTENT_HASH_MISMATCH", &object)
	}
	if err = pipeline.objects.Verify(ctx, ObjectExpected{Key: plan.ObjectKey,
		Version: objectResult.Version, Generation: plan.Generation,
		CiphertextBytes:  cipherReceipt.CiphertextBytes,
		CiphertextSHA256: cipherReceipt.CiphertextSHA256}); err != nil {
		return EffectReceipt{}, pipeline.failure(plan, FailureTransient,
			"DOCUMENT_OBJECT_VERIFY_FAILED", &object)
	}
	return EffectReceipt{Plan: plan, Object: object, DetectedKind: scanResult.DetectedKind,
		DetectedMIME: scanResult.DetectedMIME,
		ScanRevision: scanResult.SignatureRevision, ScannedAt: scanResult.ScannedAt}, nil
}

type cipherOutcome struct {
	receipt CipherReceipt
	err     error
}

func (pipeline *EffectPipeline) stream(ctx context.Context, plan AttemptPlan,
	content io.Reader, scan ScanSession, key DataKey,
) (CipherReceipt, ObjectPutResult, error) {
	reader, writer := io.Pipe()
	outcome := make(chan cipherOutcome, 1)
	go func() {
		receipt, err := pipeline.cipher.Encrypt(ctx, io.TeeReader(content, scan), writer,
			key.Plaintext, CipherMetadata{TenantID: plan.TenantID, DocumentID: plan.DocumentID,
				VersionID: plan.CandidateVersionID, Generation: plan.Generation,
				PlaintextBytes: plan.ExpectedBytes})
		_ = writer.CloseWithError(err)
		outcome <- cipherOutcome{receipt: receipt, err: err}
	}()
	expectedBytes, sizeErr := domain.CiphertextUpperBound(plan.ExpectedBytes)
	if sizeErr != nil {
		_ = reader.CloseWithError(sizeErr)
		cipherResult := <-outcome
		return cipherResult.receipt, ObjectPutResult{}, sizeErr
	}
	object, putErr := pipeline.objects.Put(ctx, ObjectPutRequest{Key: plan.ObjectKey,
		Generation: plan.Generation, CiphertextBytes: expectedBytes, Body: reader})
	_ = reader.CloseWithError(putErr)
	cipherResult := <-outcome
	if cipherResult.err != nil {
		return CipherReceipt{}, ObjectPutResult{}, ErrUnavailable
	}
	if putErr != nil {
		object, putErr = pipeline.objects.Reconcile(ctx, ObjectExpected{Key: plan.ObjectKey,
			Generation: plan.Generation, CiphertextBytes: cipherResult.receipt.CiphertextBytes,
			CiphertextSHA256: cipherResult.receipt.CiphertextSHA256})
	}
	if putErr != nil || object.Version == "" || object.Generation != plan.Generation {
		return cipherResult.receipt, ObjectPutResult{}, ErrUnavailable
	}
	return cipherResult.receipt, object, nil
}

func (pipeline *EffectPipeline) failure(plan AttemptPlan, kind EffectFailureKind,
	code string, object *domain.ObjectReceipt,
) *EffectFailure {
	now := time.Time{}
	if pipeline != nil && pipeline.clock != nil {
		now = pipeline.clock.Now()
	}
	return &EffectFailure{Plan: plan, Kind: kind, StableCode: code,
		Object: object, OccurredAt: now}
}

func validAttemptPlan(plan AttemptPlan) bool {
	return plan.UploadID.UUIDv7() && plan.TenantID.Valid() && plan.ActorID.Valid() &&
		plan.DocumentID.UUIDv7() && plan.CandidateVersionID.UUIDv7() &&
		plan.BillingReservationID.UUIDv7() && plan.EffectReceiptID.UUIDv7() &&
		plan.Generation > 0 && plan.ObjectKey != "" && plan.ExpectedBytes > 0 &&
		plan.ExpectedContentSHA256 != [32]byte{} && plan.DeclaredKind != ""
}
