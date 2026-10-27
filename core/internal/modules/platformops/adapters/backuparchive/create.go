package backuparchive

import (
	"context"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/backupcrypto"
)

func (adapter *Adapter) Create(
	ctx context.Context,
	dsn string,
	snapshotID string,
	target io.Writer,
	archiveContext backupcrypto.Context,
	keys backupcrypto.SealKeys,
) (backupcrypto.Result, error) {
	if adapter == nil || ctx == nil || ctx.Err() != nil || target == nil {
		return backupcrypto.Result{}, ErrArchive
	}
	reader, writer := io.Pipe()
	result := make(chan error, 1)
	go func() {
		err := adapter.dumper.Dump(ctx, dsn, snapshotID, writer)
		_ = writer.CloseWithError(err)
		result <- err
	}()
	encrypted, encryptErr := backupcrypto.Encrypt(reader, target, archiveContext, keys)
	if encryptErr != nil {
		_ = reader.CloseWithError(encryptErr)
	} else {
		_ = reader.Close()
	}
	dumpErr := <-result
	if encryptErr != nil || dumpErr != nil {
		return backupcrypto.Result{}, ErrArchive
	}
	return encrypted, nil
}
