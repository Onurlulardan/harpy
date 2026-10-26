package s3object

import (
	"context"

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
		return aws.Credentials{}, ErrWorkloadIdentity
	}
	credentials, err := provider.provider.Retrieve(ctx)
	if err != nil {
		return aws.Credentials{}, err
	}
	if _, allowed := workloadCredentialSources[credentials.Source]; !allowed {
		return aws.Credentials{}, ErrWorkloadIdentity
	}
	return credentials, nil
}
