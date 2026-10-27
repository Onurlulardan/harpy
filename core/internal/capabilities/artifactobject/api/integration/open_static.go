//go:build integration

package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/adapters/s3object"
	artifactapi "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

type realClock struct{}

type minIOACLClient struct {
	*s3.Client
	public bool
}

func (client minIOACLClient) GetBucketAcl(ctx context.Context,
	input *s3.GetBucketAclInput,
	options ...func(*s3.Options)) (*s3.GetBucketAclOutput, error) {
	output, err := client.Client.GetBucketAcl(ctx, input, options...)
	if err == nil && output != nil && client.public {
		uri := "http://acs.amazonaws.com/groups/global/AllUsers"
		output.Grants = append(output.Grants, types.Grant{Grantee: &types.Grantee{
			Type: types.TypeGroup, URI: &uri}, Permission: types.PermissionRead})
		return output, nil
	}
	if err != nil || output == nil || output.Owner == nil || len(output.Grants) != 1 ||
		output.Grants[0].Grantee == nil ||
		output.Grants[0].Grantee.Type != types.TypeCanonicalUser ||
		output.Grants[0].Grantee.URI != nil ||
		output.Grants[0].Grantee.EmailAddress != nil {
		return output, err
	}
	owner := "minio-private-owner"
	output.Owner.ID = &owner
	output.Grants[0].Grantee.ID = &owner
	return output, nil
}

func (realClock) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

func OpenStaticS3(ctx context.Context, endpoint, region, bucket, owner,
	prefix, contract string, pathStyle bool, maximumCiphertextBytes int64,
	accessKeyID, secretKey, sessionToken string) (artifactapi.Store, error) {
	return openStaticS3(ctx, endpoint, region, bucket, owner, prefix, contract,
		pathStyle, maximumCiphertextBytes, accessKeyID, secretKey, sessionToken, false)
}

func OpenStaticS3PublicACLProbe(ctx context.Context, endpoint, region, bucket,
	owner, prefix, contract string, pathStyle bool, maximumCiphertextBytes int64,
	accessKeyID, secretKey, sessionToken string) (artifactapi.Store, error) {
	return openStaticS3(ctx, endpoint, region, bucket, owner, prefix, contract,
		pathStyle, maximumCiphertextBytes, accessKeyID, secretKey, sessionToken, true)
}

func openStaticS3(ctx context.Context, endpoint, region, bucket, owner,
	prefix, contract string, pathStyle bool, maximumCiphertextBytes int64,
	accessKeyID, secretKey, sessionToken string, public bool) (artifactapi.Store, error) {
	client, err := s3object.NewStaticCredentialsClient(ctx, endpoint, region,
		pathStyle, accessKeyID, secretKey, sessionToken)
	if err != nil {
		return nil, err
	}
	store, err := s3object.NewVerified(ctx, minIOACLClient{Client: client, public: public},
		s3object.Config{
			Bucket: bucket, ExpectedBucketOwner: owner, ObjectPrefix: prefix,
			Contract: contract, MaximumCiphertextBytes: maximumCiphertextBytes,
		}, realClock{})
	if err != nil {
		return nil, fmt.Errorf("verify integration artifact S3 store: %w", err)
	}
	return store, nil
}
