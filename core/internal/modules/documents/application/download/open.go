package download

import (
	"context"
	"io"

	copyapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/copy"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Open(ctx context.Context, runner ScopeRunner,
	request Request,
) (Result, error) {
	operation := OperationCurrent
	if request.VersionID.Valid() {
		operation = OperationVersion
	}
	return application.open(ctx, runner, request, operation, Permission)
}

func (application *Application) OpenForCopy(ctx context.Context, runner ScopeRunner,
	request Request,
) (Result, error) {
	return application.open(ctx, runner, request, copyapp.OperationCopy,
		copyapp.PermissionCopy)
}

func (application *Application) open(ctx context.Context, runner ScopeRunner,
	request Request, operation, permission string,
) (Result, error) {
	if application == nil || foundation.NilPort(runner) || !request.ItemID.UUIDv7() ||
		request.VersionID.Valid() && !request.VersionID.UUIDv7() ||
		!platformtx.NetworkAllowed(ctx) {
		return Result{}, ErrInvalid
	}
	descriptor, err := application.descriptor(ctx, runner, request, operation, permission)
	if err != nil || !validDescriptor(descriptor, request) {
		if err != nil {
			return Result{}, err
		}
		return Result{}, ErrIntegrity
	}
	key, err := application.keyring.Unwrap(ctx, descriptor)
	if err != nil {
		return Result{}, ErrUnavailable
	}
	expected := ObjectExpected{Key: descriptor.Object.Key,
		Version: descriptor.Object.Version, Generation: descriptor.Object.Generation,
		CiphertextBytes:  descriptor.Object.CiphertextBytes,
		CiphertextSHA256: descriptor.Object.CiphertextSHA256}
	object, err := application.objects.OpenExact(ctx, expected)
	if err != nil {
		clear(key[:])
		return Result{}, mapError(err)
	}
	content := application.decryptStream(ctx, object, key, descriptor)
	return Result{ItemID: descriptor.ItemID, VersionID: descriptor.VersionID,
		Filename: descriptor.OriginalName, MIME: descriptor.MIME,
		Bytes:  descriptor.Object.PlaintextBytes,
		SHA256: descriptor.Object.PlaintextSHA256, Content: content}, nil
}

func validDescriptor(value Descriptor, request Request) bool {
	expectedKey, err := domain.ContentObjectKey(value.TenantID, value.ItemID, value.VersionID)
	return err == nil && value.ItemID == request.ItemID && value.VersionID.UUIDv7() &&
		(!request.VersionID.Valid() || value.VersionID == request.VersionID) &&
		value.OriginalName != "" && value.MIME != "" && value.Object.Key == expectedKey &&
		value.Object.Version != "" && value.Object.Generation > 0 &&
		value.Object.PlaintextBytes > 0 &&
		value.Object.CiphertextBytes > value.Object.PlaintextBytes &&
		value.Object.PlaintextSHA256 != [32]byte{} &&
		value.Object.CiphertextSHA256 != [32]byte{} && len(value.Object.WrappedDEK) >= 115
}

func (application *Application) decryptStream(ctx context.Context, source io.ReadCloser,
	key [32]byte, descriptor Descriptor,
) io.ReadCloser {
	streamContext, cancel := context.WithCancel(ctx)
	reader, writer := io.Pipe()
	stream := &contentStream{PipeReader: reader, source: source, cancel: cancel}
	go func() {
		defer clear(key[:])
		defer source.Close()
		receipt, err := application.cipher.Decrypt(streamContext, source, writer, key,
			CipherMetadata{TenantID: descriptor.TenantID, ItemID: descriptor.ItemID,
				VersionID: descriptor.VersionID, Generation: descriptor.Object.Generation,
				PlaintextBytes: descriptor.Object.PlaintextBytes})
		if err == nil && (receipt.PlaintextSHA256 != descriptor.Object.PlaintextSHA256 ||
			receipt.CiphertextSHA256 != descriptor.Object.CiphertextSHA256) {
			err = ErrIntegrity
		}
		_ = writer.CloseWithError(err)
	}()
	return stream
}
