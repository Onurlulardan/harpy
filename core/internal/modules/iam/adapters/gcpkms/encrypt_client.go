package gcpkms

import (
	"context"

	"cloud.google.com/go/auth/credentials"
	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/option"
)

type rawEncryptClient interface {
	RawEncrypt(
		context.Context,
		*kmspb.RawEncryptRequest,
		...gax.CallOption,
	) (*kmspb.RawEncryptResponse, error)
	Close() error
}

type encryptClientOpener func(context.Context) (rawEncryptClient, error)

func openGoogleEncryptClient(ctx context.Context) (rawEncryptClient, error) {
	identity, err := credentials.DetectDefault(&credentials.DetectOptions{
		Scopes: []string{"https://www.googleapis.com/auth/cloudkms"},
	})
	if err != nil || identity == nil ||
		!workloadCredentialDocumentAllowed(identity.JSON()) {
		return nil, ErrUnavailable
	}
	client, err := kms.NewKeyManagementClient(
		ctx, option.WithAuthCredentials(identity),
	)
	if err != nil {
		return nil, ErrUnavailable
	}
	return client, nil
}
