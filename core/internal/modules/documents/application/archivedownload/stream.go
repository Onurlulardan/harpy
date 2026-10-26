package archivedownload

import (
	"context"
	"io"
	"sync"

	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
)

func (application *Application) decrypt(ctx context.Context, source io.ReadCloser,
	key [32]byte, descriptor downloadapp.Descriptor,
) io.ReadCloser {
	streamContext, cancel := context.WithCancel(ctx)
	reader, writer := io.Pipe()
	stream := &archiveStream{PipeReader: reader, source: source, cancel: cancel}
	go func() {
		defer clear(key[:])
		defer source.Close()
		receipt, err := application.cipher.Decrypt(streamContext, source, writer, key,
			downloadapp.CipherMetadata{TenantID: descriptor.TenantID,
				ItemID: descriptor.ItemID, VersionID: descriptor.VersionID,
				Generation:     descriptor.Object.Generation,
				PlaintextBytes: descriptor.Object.PlaintextBytes})
		if err == nil && (receipt.PlaintextSHA256 != descriptor.Object.PlaintextSHA256 ||
			receipt.CiphertextSHA256 != descriptor.Object.CiphertextSHA256) {
			err = ErrIntegrity
		}
		_ = writer.CloseWithError(err)
	}()
	return stream
}

type archiveStream struct {
	*io.PipeReader
	source io.Closer
	cancel context.CancelFunc
	once   sync.Once
}

func (stream *archiveStream) Close() error {
	var result error
	stream.once.Do(func() {
		stream.cancel()
		result = stream.PipeReader.Close()
		if err := stream.source.Close(); result == nil {
			result = err
		}
	})
	return result
}
