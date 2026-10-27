package recoverys3

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	environmentrecoveryv1 "github.com/Onurlulardan/harpy/contracts/environmentrecovery/v1"
	recovery "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application/recoveryorchestration"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoverystorage"
)

type databaseRestoreStorage struct {
	manifest        recoverystorage.Manifest
	recoveryClient  *s3.Client
	recoveryReader  *Reader
	rehydrateClient *s3.Client
}

func LoadDatabaseRestoreStorage(
	ctx context.Context,
	path string,
	environment string,
	read ArtifactReader,
) (databaseRestoreStorage, error) {
	if read == nil {
		return databaseRestoreStorage{}, fmt.Errorf("recovery artifact reader is unavailable")
	}
	document, err := read(
		path, recoverystorage.MaximumManifestBytes, false,
	)
	if err != nil {
		return databaseRestoreStorage{}, err
	}
	manifest, err := recoverystorage.Parse(document)
	if err != nil || !manifest.AllowsProtected(environment) {
		return databaseRestoreStorage{},
			fmt.Errorf("recovery storage policy is invalid")
	}
	restoreCredential, err := loadRecoveryCredential(
		manifest.Credentials.RestoreRead, read,
	)
	if err != nil {
		return databaseRestoreStorage{}, err
	}
	recoveryClient, err := NewStaticClient(
		ctx, manifest.Endpoint, manifest.Region, manifest.PathStyle,
		manifest.TLSCASHA256, restoreCredential,
	)
	if err != nil {
		return databaseRestoreStorage{}, err
	}
	if _, err = Preflight(ctx, recoveryClient, manifest); err != nil {
		return databaseRestoreStorage{}, err
	}
	rehydrateCredential, err := loadRecoveryCredential(
		manifest.Credentials.DocumentsRehydrate, read,
	)
	if err != nil {
		return databaseRestoreStorage{}, err
	}
	target := manifest.DocumentsRehydrateTarget
	rehydrateClient, err := NewStaticClient(
		ctx, target.Endpoint, target.Region, target.PathStyle,
		target.TLSCASHA256, rehydrateCredential,
	)
	if err != nil ||
		PreflightVersionedPrivate(
			ctx, rehydrateClient, target,
		) != nil {
		return databaseRestoreStorage{}, fmt.Errorf(
			"documents rehydrate target is unavailable",
		)
	}
	return databaseRestoreStorage{
		manifest: manifest, recoveryClient: recoveryClient,
		recoveryReader: NewReader(
			recoveryClient, manifest.Bucket, manifest.ExpectedOwner,
		),
		rehydrateClient: rehydrateClient,
	}, nil
}

func (storage databaseRestoreStorage) rehydrateWriter(
	operationID string,
) *RehydrateWriter {
	target := storage.manifest.DocumentsRehydrateTarget
	return NewRehydrateWriter(
		storage.rehydrateClient, target.Bucket, target.ExpectedOwner, operationID,
	)
}

func (storage databaseRestoreStorage) Prefix(
	environmentID, recoverySetID string,
) (string, error) {
	return storage.manifest.Prefix(environmentID, recoverySetID)
}

func (storage databaseRestoreStorage) RecoveryKeyProvider() (string, string) {
	provider := storage.manifest.RecoveryKeyProvider
	return provider.Kind, provider.Handle
}

func (storage databaseRestoreStorage) RecoveryReader() recovery.RecoveryObjectReader {
	return storage.recoveryReader
}

func (storage databaseRestoreStorage) RehydrateWriter(
	operationID string,
) recovery.RecoveryObjectWriter {
	return storage.rehydrateWriter(operationID)
}

func (storage databaseRestoreStorage) ManifestVersions(
	ctx context.Context,
	environmentID, recoverySetID string,
) ([][]byte, error) {
	return readManifestVersions(ctx, storage, environmentID, recoverySetID)
}

func readManifestVersions(
	ctx context.Context,
	storage databaseRestoreStorage,
	environmentID, recoverySetID string,
) ([][]byte, error) {
	prefix, err := storage.manifest.Prefix(environmentID, recoverySetID)
	if err != nil {
		return nil, err
	}
	key := prefix + "manifest.json"
	output, err := storage.recoveryClient.ListObjectVersions(
		ctx, &s3.ListObjectVersionsInput{
			Bucket: aws.String(storage.manifest.Bucket),
			Prefix: aws.String(key), MaxKeys: aws.Int32(1000),
			ExpectedBucketOwner: aws.String(storage.manifest.ExpectedOwner),
		},
	)
	if err != nil || output == nil ||
		output.IsTruncated != nil && *output.IsTruncated ||
		len(output.DeleteMarkers) != 0 {
		return nil, fmt.Errorf("recovery manifest inventory is unavailable")
	}
	documents := make([][]byte, 0, len(output.Versions))
	for _, version := range output.Versions {
		if version.Key == nil || *version.Key != key || version.VersionId == nil {
			continue
		}
		value, readErr := storage.recoveryClient.GetObject(
			ctx, &s3.GetObjectInput{
				Bucket: aws.String(storage.manifest.Bucket), Key: aws.String(key),
				VersionId:           version.VersionId,
				ExpectedBucketOwner: aws.String(storage.manifest.ExpectedOwner),
			},
		)
		if readErr != nil || value == nil || value.Body == nil {
			return nil, fmt.Errorf("recovery manifest is unavailable")
		}
		document, readErr := io.ReadAll(io.LimitReader(
			value.Body, environmentrecoveryv1.MaximumArtifactBytes+1,
		))
		closeErr := value.Body.Close()
		if readErr != nil || closeErr != nil ||
			len(document) > environmentrecoveryv1.MaximumArtifactBytes {
			return nil, fmt.Errorf("recovery manifest exceeds its limit")
		}
		documents = append(documents, document)
	}
	return documents, nil
}
