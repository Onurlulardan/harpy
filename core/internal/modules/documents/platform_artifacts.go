package documents

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/cipher"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/keyring"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/s3blob"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type platformArtifacts struct {
	objects *s3blob.Store
	keyring *keyring.Ring
	cipher  *cipher.Runtime
}

func openPlatformArtifacts(ctx context.Context,
	settings config.DocumentsArtifacts,
) (platformArtifacts, error) {
	if ctx == nil || ctx.Err() != nil || settings.Validate() != nil {
		return platformArtifacts{}, fmt.Errorf("Documents artifact settings are invalid")
	}
	storage, err := settings.Storage()
	if err != nil {
		return platformArtifacts{}, err
	}
	client, bucket, owner, err := openPlatformS3Client(ctx, storage)
	if err != nil {
		return platformArtifacts{}, err
	}
	objects, err := s3blob.NewVerified(ctx, client,
		s3blob.Config{Bucket: bucket, ExpectedBucketOwner: owner}, platformClock{})
	if err != nil {
		return platformArtifacts{}, fmt.Errorf("verify Documents S3 target: %w", err)
	}
	configured, err := settings.KEKKeyring()
	if err != nil {
		return platformArtifacts{}, err
	}
	releases, err := configured.Releases()
	if err != nil {
		return platformArtifacts{}, err
	}
	keys := make([]keyring.Release, 0, len(releases))
	for _, release := range releases {
		id, material, current, valueErr := release.Values()
		if valueErr != nil {
			return platformArtifacts{}, valueErr
		}
		keys = append(keys, keyring.Release{ID: id, Key: material, Current: current})
	}
	ring, err := keyring.New(keys, rand.Reader)
	if err != nil {
		return platformArtifacts{}, fmt.Errorf("build Documents keyring: %w", err)
	}
	cryptoRuntime, err := cipher.New(rand.Reader)
	if err != nil {
		return platformArtifacts{}, fmt.Errorf("build Documents cipher: %w", err)
	}
	return platformArtifacts{objects: objects, keyring: ring,
		cipher: cryptoRuntime}, nil
}
