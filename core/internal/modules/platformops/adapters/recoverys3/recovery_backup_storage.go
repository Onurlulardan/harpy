package recoverys3

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	recovery "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/recoveryorchestration"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoverystorage"
)

type ArtifactReader func(string, int64, bool) ([]byte, error)

type DatabaseBackupStorage struct {
	manifest        recoverystorage.Manifest
	recoveryClient  *s3.Client
	recoveryReader  *Reader
	documentsReader *Reader
}

func LoadDatabaseBackupStorage(
	ctx context.Context,
	path string,
	environment string,
	read ArtifactReader,
) (DatabaseBackupStorage, error) {
	if read == nil {
		return DatabaseBackupStorage{}, fmt.Errorf("recovery artifact reader is unavailable")
	}
	document, err := read(
		path, recoverystorage.MaximumManifestBytes, false,
	)
	if err != nil {
		return DatabaseBackupStorage{}, err
	}
	manifest, err := recoverystorage.Parse(document)
	if err != nil || !manifest.AllowsProtected(environment) {
		return DatabaseBackupStorage{}, fmt.Errorf("recovery storage policy is invalid")
	}
	recoveryCredential, err := loadRecoveryCredential(
		manifest.Credentials.RecoveryWrite, read,
	)
	if err != nil {
		return DatabaseBackupStorage{}, err
	}
	recoveryClient, err := NewStaticClient(
		ctx, manifest.Endpoint, manifest.Region, manifest.PathStyle,
		manifest.TLSCASHA256, recoveryCredential,
	)
	if err != nil {
		return DatabaseBackupStorage{}, err
	}
	if _, err = Preflight(ctx, recoveryClient, manifest); err != nil {
		return DatabaseBackupStorage{}, err
	}
	sourceCredential, err := loadRecoveryCredential(
		manifest.Credentials.DocumentsSourceRead, read,
	)
	if err != nil {
		return DatabaseBackupStorage{}, err
	}
	source := manifest.DocumentsSource
	sourceClient, err := NewStaticClient(
		ctx, source.Endpoint, source.Region, source.PathStyle,
		source.TLSCASHA256, sourceCredential,
	)
	if err != nil {
		return DatabaseBackupStorage{}, err
	}
	return DatabaseBackupStorage{
		manifest: manifest, recoveryClient: recoveryClient,
		recoveryReader: NewReader(
			recoveryClient, manifest.Bucket, manifest.ExpectedOwner,
		),
		documentsReader: NewReader(
			sourceClient, source.Bucket, source.ExpectedOwner,
		),
	}, nil
}

func loadRecoveryCredential(
	path string,
	read ArtifactReader,
) (Credential, error) {
	document, err := read(
		path, MaximumCredentialBytes, true,
	)
	if err != nil {
		return Credential{}, err
	}
	credential, parseErr := ParseCredential(document)
	clear(document)
	return credential, parseErr
}

func (storage DatabaseBackupStorage) Prefix(
	environmentID, recoverySetID string,
) (string, error) {
	return storage.manifest.Prefix(environmentID, recoverySetID)
}

func (storage DatabaseBackupStorage) RetentionDays() int {
	return storage.manifest.ObjectLock.RetentionDays
}

func (storage DatabaseBackupStorage) RecoveryKeyProvider() (string, string) {
	provider := storage.manifest.RecoveryKeyProvider
	return provider.Kind, provider.Handle
}

func (storage DatabaseBackupStorage) RecoveryReader() recovery.RecoveryObjectReader {
	return storage.recoveryReader
}

func (storage DatabaseBackupStorage) DocumentsReader() recovery.RecoveryObjectReader {
	return storage.documentsReader
}

func (storage DatabaseBackupStorage) Writer(now time.Time) recovery.RecoveryObjectWriter {
	return NewWriter(
		storage.recoveryClient, storage.manifest.Bucket,
		storage.manifest.ExpectedOwner,
		storage.manifest.ObjectLock.RetentionDays, now,
	)
}
