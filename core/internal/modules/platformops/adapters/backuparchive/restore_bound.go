package backuparchive

import (
	"context"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/backupcrypto"
)

func (adapter *Adapter) RestoreBound(
	ctx context.Context,
	dsn string,
	source io.Reader,
	expected backupcrypto.BoundContext,
	resolver backupcrypto.OpenKeyResolver,
) (int64, backupcrypto.Context, error) {
	if adapter == nil || ctx == nil || ctx.Err() != nil ||
		source == nil || resolver == nil {
		return 0, backupcrypto.Context{}, ErrArchive
	}
	reader, writer := io.Pipe()
	type decryptResult struct {
		bytes   int64
		binding backupcrypto.Context
		err     error
	}
	result := make(chan decryptResult, 1)
	go func() {
		count, binding, err := backupcrypto.DecryptBound(
			source, writer, expected, resolver,
		)
		_ = writer.CloseWithError(err)
		result <- decryptResult{bytes: count, binding: binding, err: err}
	}()
	restoreErr := adapter.restorer.Restore(ctx, dsn, reader)
	if restoreErr != nil {
		_ = reader.CloseWithError(restoreErr)
	} else {
		_ = reader.Close()
	}
	decrypted := <-result
	if restoreErr != nil || decrypted.err != nil {
		return 0, backupcrypto.Context{}, ErrArchive
	}
	return decrypted.bytes, decrypted.binding, nil
}
