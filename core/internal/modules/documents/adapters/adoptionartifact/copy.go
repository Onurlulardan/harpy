package adoptionartifact

import (
	"context"
	"io"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (copier *Copier) Copy(ctx context.Context,
	authority adoptionapp.Authority, plan adoptionapp.Checkpoint,
) (adoptionapp.CopyReceipt, error) {
	if copier == nil || ctx == nil || ctx.Err() != nil ||
		!foundation.ExternalEffectsAllowed(ctx) || !authority.TenantID.Valid() ||
		authority.RunID != plan.RunID || plan.State != adoptionapp.CheckpointCopying ||
		plan.SourceBytes < 1 || plan.SourceSHA256 == [32]byte{} ||
		plan.SourceMetadataSHA256 == [32]byte{} || plan.SourceLSN == "" {
		return adoptionapp.CopyReceipt{}, adoptionapp.ErrInvalid
	}
	metadata, err := copier.metadata.VerifyInactive(ctx, plan)
	if err != nil || metadata != plan.SourceMetadataSHA256 {
		return adoptionapp.CopyReceipt{}, adoptionapp.ErrBlocked
	}
	source, err := copier.source.OpenExact(ctx, plan.SourceObjectKey, plan.SourceRevision)
	if err != nil || source == nil {
		return adoptionapp.CopyReceipt{}, adoptionapp.ErrUnavailable
	}
	defer source.Close()
	key, err := copier.keyring.CreateDataKey(ctx, authority.TenantID,
		plan.TargetItemID, plan.TargetVersionID)
	if err != nil || key.KeyID == "" || len(key.Wrapped) < 115 {
		clear(key.Plaintext[:])
		return adoptionapp.CopyReceipt{}, adoptionapp.ErrUnavailable
	}
	defer clear(key.Plaintext[:])
	objectKey, err := domain.ContentObjectKey(authority.TenantID,
		plan.TargetItemID, plan.TargetVersionID)
	if err != nil {
		return adoptionapp.CopyReceipt{}, adoptionapp.ErrInvalid
	}
	cipherReceipt, object, err := copier.stream(ctx, source, authority,
		plan, objectKey, key.Plaintext)
	if err != nil || cipherReceipt.PlaintextSHA256 != plan.SourceSHA256 {
		return adoptionapp.CopyReceipt{}, adoptionapp.ErrUnavailable
	}
	return adoptionapp.CopyReceipt{TargetObject: domain.ObjectReceipt{Key: objectKey,
		Version: object.Version, KeyID: key.KeyID, Generation: object.Generation,
		PlaintextBytes: plan.SourceBytes, CiphertextBytes: cipherReceipt.CiphertextBytes,
		PlaintextSHA256:  cipherReceipt.PlaintextSHA256,
		CiphertextSHA256: cipherReceipt.CiphertextSHA256,
		WrappedDEK:       append([]byte(nil), key.Wrapped...)}, MetadataSHA256: metadata,
		AppliedLSN: plan.SourceLSN}, nil
}

func (copier *Copier) stream(ctx context.Context, source io.Reader,
	authority adoptionapp.Authority, plan adoptionapp.Checkpoint,
	objectKey string, key [32]byte,
) (uploadapp.CipherReceipt, uploadapp.ObjectPutResult, error) {
	reader, writer := io.Pipe()
	outcome := make(chan cipherOutcome, 1)
	go func() {
		receipt, err := copier.cipher.Encrypt(ctx, source, writer, key,
			uploadapp.CipherMetadata{TenantID: authority.TenantID,
				DocumentID: plan.TargetItemID,
				VersionID:  plan.TargetVersionID, Generation: 1,
				PlaintextBytes: plan.SourceBytes})
		_ = writer.CloseWithError(err)
		outcome <- cipherOutcome{receipt: receipt, err: err}
	}()
	bound, err := domain.CiphertextUpperBound(plan.SourceBytes)
	if err != nil {
		_ = reader.CloseWithError(err)
		return uploadapp.CipherReceipt{}, uploadapp.ObjectPutResult{}, err
	}
	object, putErr := copier.objects.Put(ctx, uploadapp.ObjectPutRequest{Key: objectKey,
		Generation: 1, CiphertextBytes: bound, Body: reader})
	_ = reader.CloseWithError(putErr)
	ciphered := <-outcome
	if ciphered.err != nil {
		return uploadapp.CipherReceipt{}, uploadapp.ObjectPutResult{}, ciphered.err
	}
	if putErr != nil {
		object, putErr = copier.objects.Reconcile(ctx, uploadapp.ObjectExpected{Key: objectKey,
			Generation: 1, CiphertextBytes: ciphered.receipt.CiphertextBytes,
			CiphertextSHA256: ciphered.receipt.CiphertextSHA256})
	}
	if putErr != nil || object.Version == "" {
		return uploadapp.CipherReceipt{}, uploadapp.ObjectPutResult{}, adoptionapp.ErrUnavailable
	}
	return ciphered.receipt, object, copier.objects.Verify(ctx,
		uploadapp.ObjectExpected{Key: objectKey, Version: object.Version, Generation: 1,
			CiphertextBytes:  ciphered.receipt.CiphertextBytes,
			CiphertextSHA256: ciphered.receipt.CiphertextSHA256})
}

type cipherOutcome struct {
	receipt uploadapp.CipherReceipt
	err     error
}
