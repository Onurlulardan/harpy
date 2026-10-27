package cipher

import (
	"context"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/blobcrypto"
	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
)

func (runtime *Runtime) Decrypt(ctx context.Context, ciphertext io.Reader,
	plaintext io.Writer, key [32]byte, metadata downloadapp.CipherMetadata,
) (downloadapp.CipherReceipt, error) {
	if runtime == nil || ctx == nil || ctx.Err() != nil {
		return downloadapp.CipherReceipt{}, downloadapp.ErrUnavailable
	}
	receipt, err := blobcrypto.Decrypt(ciphertext, plaintext, key[:], blobcrypto.Metadata{
		TenantID: metadata.TenantID, DocumentID: metadata.ItemID,
		VersionID: metadata.VersionID, Generation: metadata.Generation,
		PlaintextBytes: metadata.PlaintextBytes})
	if err != nil {
		return downloadapp.CipherReceipt{}, err
	}
	return downloadapp.CipherReceipt{PlaintextSHA256: receipt.PlaintextSHA256,
		CiphertextSHA256: receipt.CiphertextSHA256}, nil
}

var _ downloadapp.Cipher = (*Runtime)(nil)
