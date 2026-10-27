package artifactobject

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/adapters/s3object"
	artifactapi "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

func OpenS3(ctx context.Context, settings config.ArtifactS3,
	prefix, contract string, maximumCiphertextBytes int64) (artifactapi.Store, error) {
	if ctx == nil || ctx.Err() != nil || settings.Validate() != nil {
		return nil, fmt.Errorf("artifact S3 settings are invalid")
	}
	endpoint, region, bucket, owner, pathStyle, err := settings.Connection()
	if err != nil {
		return nil, err
	}
	client, err := openS3Client(ctx, settings, endpoint, region, pathStyle)
	if err != nil {
		return nil, err
	}
	return openVerifiedS3(ctx, client, bucket, owner, prefix, contract,
		maximumCiphertextBytes)
}

func openVerifiedS3(ctx context.Context, client s3object.API,
	bucket, owner, prefix, contract string,
	maximumCiphertextBytes int64) (artifactapi.Store, error) {
	store, err := s3object.NewVerified(ctx, client, s3object.Config{
		Bucket: bucket, ExpectedBucketOwner: owner, ObjectPrefix: prefix,
		Contract: contract, MaximumCiphertextBytes: maximumCiphertextBytes,
	}, realClock{})
	if err != nil {
		return nil, fmt.Errorf("verify artifact S3 store: %w", err)
	}
	return store, nil
}

func openS3Client(ctx context.Context, settings config.ArtifactS3,
	endpoint, region string, pathStyle bool) (s3object.API, error) {
	credentials, err := settings.Credentials()
	if err != nil {
		return nil, err
	}
	source, err := credentials.Source()
	if err != nil {
		return nil, err
	}
	if source == config.ArtifactWorkloadIdentity {
		return s3object.NewWorkloadIdentityClient(ctx, endpoint, region, pathStyle)
	}
	access, secret, token, _, err := credentials.Static()
	if err != nil {
		return nil, err
	}
	return s3object.NewStaticCredentialsClient(ctx, endpoint, region, pathStyle,
		access.Value(), secret.Value(), token.Value())
}
