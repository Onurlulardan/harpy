package archiveartifact

import (
	"context"
	"io"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type cipherOutcome struct {
	receipt uploadapp.CipherReceipt
	err     error
}

func (writer *Writer) stream(ctx context.Context, plan archiveapp.Plan,
	refresh func(int) error, plaintextBytes int64, objectKey string, key [32]byte,
) (uploadapp.CipherReceipt, uploadapp.ObjectPutResult, error) {
	plainReader, plainWriter := io.Pipe()
	cipherReader, cipherWriter := io.Pipe()
	zipOutcome := make(chan error, 1)
	cipherDone := make(chan cipherOutcome, 1)
	go func() {
		err := writeZIP(ctx, plainWriter, plan.Entries, writer.contents, refresh)
		_ = plainWriter.CloseWithError(err)
		zipOutcome <- err
	}()
	go func() {
		receipt, err := writer.cipher.Encrypt(ctx, plainReader, cipherWriter, key,
			uploadapp.CipherMetadata{TenantID: plan.TenantID, DocumentID: plan.ID,
				VersionID: plan.ID, Generation: 1, PlaintextBytes: plaintextBytes})
		_ = plainReader.CloseWithError(err)
		_ = cipherWriter.CloseWithError(err)
		cipherDone <- cipherOutcome{receipt: receipt, err: err}
	}()
	ciphertextBytes, err := archiveCiphertextBytes(plaintextBytes)
	if err != nil {
		_ = cipherReader.CloseWithError(err)
		return uploadapp.CipherReceipt{}, uploadapp.ObjectPutResult{}, err
	}
	object, putErr := writer.objects.Put(ctx, uploadapp.ObjectPutRequest{Key: objectKey,
		Generation: 1, CiphertextBytes: ciphertextBytes, Body: cipherReader})
	_ = cipherReader.CloseWithError(putErr)
	cipherResult := <-cipherDone
	zipErr := <-zipOutcome
	if zipErr != nil || cipherResult.err != nil {
		return uploadapp.CipherReceipt{}, uploadapp.ObjectPutResult{}, archiveapp.ErrUnavailable
	}
	expected := uploadapp.ObjectExpected{Key: objectKey, Generation: 1,
		CiphertextBytes:  cipherResult.receipt.CiphertextBytes,
		CiphertextSHA256: cipherResult.receipt.CiphertextSHA256}
	if putErr != nil {
		object, putErr = writer.objects.Reconcile(ctx, expected)
	}
	if putErr != nil || object.Version == "" || object.Generation != 1 {
		return uploadapp.CipherReceipt{}, uploadapp.ObjectPutResult{}, archiveapp.ErrUnavailable
	}
	expected.Version = object.Version
	if err = writer.objects.Verify(ctx, expected); err != nil {
		return uploadapp.CipherReceipt{}, uploadapp.ObjectPutResult{}, archiveapp.ErrUnavailable
	}
	return cipherResult.receipt, object, nil
}

func archiveCiphertextBytes(plaintext int64) (int64, error) {
	return domain.ArchiveCiphertextUpperBound(plaintext)
}
