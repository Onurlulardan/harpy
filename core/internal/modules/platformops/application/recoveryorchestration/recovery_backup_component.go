package recoveryorchestration

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/backupcrypto"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type databaseBackupComponent struct {
	kind     string
	identity ports.ObjectIdentity
}

func (component databaseBackupComponent) manifestIdentity() recoveryset.ComponentIdentity {
	return recoveryset.ComponentIdentity{
		ObjectKey:        component.identity.Key,
		ObjectVersionID:  component.identity.VersionID,
		CiphertextSHA256: component.identity.SHA256,
		CiphertextBytes:  component.identity.Bytes,
	}
}

func putEncryptedDocument(
	ctx context.Context,
	writer RecoveryObjectWriter,
	key string,
	plaintext []byte,
	archiveContext backupcrypto.Context,
	keys backupcrypto.SealKeys,
) (databaseBackupComponent, error) {
	var encrypted bytes.Buffer
	result, err := backupcrypto.Encrypt(
		bytes.NewReader(plaintext), &encrypted, archiveContext, keys,
	)
	if err != nil {
		return databaseBackupComponent{}, err
	}
	evidence, err := writer.PutVerified(ctx, ports.WriteRequest{
		Key: key, Body: bytes.NewReader(encrypted.Bytes()),
		ContentLength:  result.CiphertextBytes,
		ExpectedSHA256: result.CiphertextSHA256,
		ExpectedBytes:  result.CiphertextBytes,
	})
	if err != nil {
		return databaseBackupComponent{}, err
	}
	return databaseBackupComponent{identity: evidence.ObjectIdentity}, nil
}

func putDatabaseArchive(
	ctx context.Context,
	writer RecoveryObjectWriter,
	key, dsn, snapshotID string,
	archive DatabaseArchive,
	archiveContext backupcrypto.Context,
	keys backupcrypto.SealKeys,
) (databaseBackupComponent, error) {
	reader, pipeWriter := io.Pipe()
	results := make(chan DatabaseArchiveResult, 1)
	go func() {
		result, err := archive.Create(
			ctx, dsn, snapshotID, pipeWriter, archiveContext, keys,
		)
		_ = pipeWriter.CloseWithError(err)
		results <- DatabaseArchiveResult{result: result, err: err}
	}()
	evidence, writeErr := writer.PutVerified(ctx, ports.WriteRequest{
		Key: key, Body: reader, ContentLength: -1,
	})
	if writeErr != nil {
		_ = reader.CloseWithError(writeErr)
	} else {
		_ = reader.Close()
	}
	archiveResult := <-results
	if writeErr != nil || archiveResult.err != nil ||
		evidence.SHA256 != archiveResult.result.CiphertextSHA256 ||
		evidence.Bytes != archiveResult.result.CiphertextBytes {
		return databaseBackupComponent{},
			fmt.Errorf("database archive write verification failed")
	}
	return databaseBackupComponent{identity: evidence.ObjectIdentity}, nil
}

type DatabaseArchiveResult struct {
	result backupcrypto.Result
	err    error
}

func verifyBackupComponent(
	ctx context.Context,
	reader RecoveryObjectReader,
	kind string,
	component recoveryset.ComponentIdentity,
) (databaseBackupComponent, error) {
	identity := ports.ObjectIdentity{
		Key: component.ObjectKey, VersionID: component.ObjectVersionID,
		SHA256: component.CiphertextSHA256, Bytes: component.CiphertextBytes,
	}
	if err := reader.DownloadVerified(ctx, identity, io.Discard); err != nil {
		return databaseBackupComponent{}, err
	}
	return databaseBackupComponent{kind: kind, identity: identity}, nil
}
