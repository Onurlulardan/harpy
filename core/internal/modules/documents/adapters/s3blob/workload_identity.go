package s3blob

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
)

var workloadCredentialSources = map[string]struct{}{
	"WebIdentityCredentials":      {},
	"CredentialsEndpointProvider": {},
	"EC2RoleProvider":             {},
}

type workloadIdentityProvider struct{ provider aws.CredentialsProvider }

func (provider workloadIdentityProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	if provider.provider == nil {
		return aws.Credentials{}, ErrConfiguration
	}
	credentials, err := provider.provider.Retrieve(ctx)
	if err != nil {
		return aws.Credentials{}, err
	}
	if _, allowed := workloadCredentialSources[credentials.Source]; !allowed {
		return aws.Credentials{}, errors.Join(ErrConfiguration,
			errors.New("documents S3 credentials are not workload identity"))
	}
	return credentials, nil
}
