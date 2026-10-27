package archiveartifact

import (
	"context"
	"io"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

type EncryptedSource struct {
	keyring downloadapp.Keyring
	cipher  downloadapp.Cipher
	objects downloadapp.ObjectStore
}

func NewEncryptedSource(keyring downloadapp.Keyring, cipher downloadapp.Cipher,
	objects downloadapp.ObjectStore,
) (*EncryptedSource, error) {
	for _, dependency := range []any{keyring, cipher, objects} {
		if foundation.NilPort(dependency) {
			return nil, archiveapp.ErrUnavailable
		}
	}
	return &EncryptedSource{keyring: keyring, cipher: cipher, objects: objects}, nil
}

func (source *EncryptedSource) Open(ctx context.Context,
	entry archiveapp.Entry,
) (io.ReadCloser, error) {
	if source == nil || ctx == nil || ctx.Err() != nil || !validSourceEntry(entry) {
		return nil, archiveapp.ErrInvalid
	}
	expected := downloadapp.ObjectExpected{Key: entry.Object.Key,
		Version: entry.Object.Version, Generation: entry.Object.Generation,
		CiphertextBytes:  entry.Object.CiphertextBytes,
		CiphertextSHA256: entry.Object.CiphertextSHA256}
	ciphertext, err := source.objects.OpenExact(ctx, expected)
	if err != nil {
		return nil, archiveapp.ErrUnavailable
	}
	descriptor := downloadapp.Descriptor{TenantID: entry.TenantID,
		ItemID: entry.DocumentID, VersionID: entry.VersionID, Object: entry.Object}
	key, err := source.keyring.Unwrap(ctx, descriptor)
	if err != nil {
		_ = ciphertext.Close()
		return nil, archiveapp.ErrUnavailable
	}
	reader, writer := io.Pipe()
	go source.decrypt(ctx, ciphertext, writer, key, entry)
	return reader, nil
}

func (source *EncryptedSource) decrypt(ctx context.Context, ciphertext io.ReadCloser,
	plaintext *io.PipeWriter, key [32]byte, entry archiveapp.Entry,
) {
	defer clear(key[:])
	receipt, err := source.cipher.Decrypt(ctx, ciphertext, plaintext, key,
		downloadapp.CipherMetadata{TenantID: entry.TenantID, ItemID: entry.DocumentID,
			VersionID: entry.VersionID, Generation: entry.Object.Generation,
			PlaintextBytes: entry.PlaintextBytes})
	closeErr := ciphertext.Close()
	if err == nil && closeErr != nil {
		err = closeErr
	}
	if err == nil && (receipt.PlaintextSHA256 != entry.PlaintextSHA256 ||
		receipt.CiphertextSHA256 != entry.Object.CiphertextSHA256) {
		err = archiveapp.ErrUnavailable
	}
	_ = plaintext.CloseWithError(err)
}

func validSourceEntry(entry archiveapp.Entry) bool {
	object := entry.Object
	return entry.TenantID.Valid() && entry.DocumentID.UUIDv7() &&
		entry.VersionID.UUIDv7() && entry.PlaintextBytes > 0 &&
		entry.PlaintextSHA256 != [32]byte{} && object.Key != "" &&
		object.Version != "" && object.Generation > 0 &&
		object.PlaintextBytes == entry.PlaintextBytes &&
		object.PlaintextSHA256 == entry.PlaintextSHA256 &&
		object.CiphertextBytes > object.PlaintextBytes &&
		object.CiphertextSHA256 != [32]byte{} && object.KeyID != "" &&
		len(object.WrappedDEK) >= 115
}

var _ ContentSource = (*EncryptedSource)(nil)
