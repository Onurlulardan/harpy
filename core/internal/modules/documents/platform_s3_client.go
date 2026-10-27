package documents

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/s3blob"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func openPlatformS3Client(ctx context.Context, settings config.ArtifactS3) (
	s3blob.API, string, string, error,
) {
	endpoint, region, bucket, owner, pathStyle, err := settings.Connection()
	if err != nil {
		return nil, "", "", err
	}
	credentials, err := settings.Credentials()
	if err != nil {
		return nil, "", "", err
	}
	source, err := credentials.Source()
	if err != nil {
		return nil, "", "", err
	}
	if source == config.ArtifactWorkloadIdentity {
		client, openErr := s3blob.NewWorkloadIdentityClient(ctx, endpoint, region, pathStyle)
		return client, bucket, owner, openErr
	}
	access, secret, token, _, err := credentials.Static()
	if err != nil {
		return nil, "", "", err
	}
	client, err := s3blob.NewStaticCredentialsClient(ctx, endpoint, region, pathStyle,
		access.Value(), secret.Value(), token.Value())
	return client, bucket, owner, err
}
