package backuparchive

import (
	"context"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/backupcrypto"
)

func (adapter *Adapter) Restore(
	ctx context.Context,
	dsn string,
	source io.Reader,
	expected backupcrypto.Context,
	keys backupcrypto.OpenKeys,
) (int64, error) {
	if adapter == nil || ctx == nil || ctx.Err() != nil || source == nil {
		return 0, ErrArchive
	}
	reader, writer := io.Pipe()
	type decryptResult struct {
		bytes int64
		err   error
	}
	result := make(chan decryptResult, 1)
	go func() {
		count, err := backupcrypto.DecryptExpected(source, writer, expected, keys)
		_ = writer.CloseWithError(err)
		result <- decryptResult{bytes: count, err: err}
	}()
	restoreErr := adapter.restorer.Restore(ctx, dsn, reader)
	if restoreErr != nil {
		_ = reader.CloseWithError(restoreErr)
	} else {
		_ = reader.Close()
	}
	decrypted := <-result
	if restoreErr != nil || decrypted.err != nil {
		return 0, ErrArchive
	}
	return decrypted.bytes, nil
}
