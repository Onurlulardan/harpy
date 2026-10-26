package cipher

import (
	"context"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/blobcrypto"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type Runtime struct{ randomness io.Reader }

func New(randomness io.Reader) (*Runtime, error) {
	if randomness == nil {
		return nil, uploadapp.ErrUnavailable
	}
	return &Runtime{randomness: randomness}, nil
}

func (runtime *Runtime) Encrypt(ctx context.Context, plaintext io.Reader,
	ciphertext io.Writer, key [32]byte, metadata uploadapp.CipherMetadata,
) (uploadapp.CipherReceipt, error) {
	if runtime == nil || runtime.randomness == nil || ctx == nil || ctx.Err() != nil {
		return uploadapp.CipherReceipt{}, uploadapp.ErrUnavailable
	}
	receipt, err := blobcrypto.Encrypt(plaintext, ciphertext, key[:], blobcrypto.Metadata{
		TenantID: metadata.TenantID, DocumentID: metadata.DocumentID,
		VersionID: metadata.VersionID, Generation: metadata.Generation,
		PlaintextBytes: metadata.PlaintextBytes}, runtime.randomness)
	if err != nil {
		return uploadapp.CipherReceipt{}, err
	}
	bytes, err := domain.CiphertextUpperBound(metadata.PlaintextBytes)
	if err != nil && metadata.PlaintextBytes > domain.MaxUploadBytes {
		bytes, err = domain.ArchiveCiphertextUpperBound(metadata.PlaintextBytes)
	}
	if err != nil {
		return uploadapp.CipherReceipt{}, err
	}
	return uploadapp.CipherReceipt{PlaintextSHA256: receipt.PlaintextSHA256,
		CiphertextSHA256: receipt.CiphertextSHA256, CiphertextBytes: bytes}, nil
}

var _ uploadapp.Cipher = (*Runtime)(nil)
