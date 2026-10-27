package recoverys3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoverystorage"
)

type VersionedAPI interface {
	GetBucketAcl(
		context.Context, *s3.GetBucketAclInput, ...func(*s3.Options),
	) (*s3.GetBucketAclOutput, error)
	GetBucketVersioning(
		context.Context, *s3.GetBucketVersioningInput, ...func(*s3.Options),
	) (*s3.GetBucketVersioningOutput, error)
}

func PreflightVersionedPrivate(
	ctx context.Context,
	client VersionedAPI,
	profile recoverystorage.StorageProfile,
) error {
	if ctx == nil || ctx.Err() != nil || client == nil ||
		profile.Bucket == "" || profile.ExpectedOwner == "" {
		return ErrPreflight
	}
	acl, err := client.GetBucketAcl(ctx, &s3.GetBucketAclInput{
		Bucket:              aws.String(profile.Bucket),
		ExpectedBucketOwner: aws.String(profile.ExpectedOwner),
	})
	if err != nil || !privateACL(acl, profile.ExpectedOwner) {
		return ErrPreflight
	}
	versioning, err := client.GetBucketVersioning(
		ctx, &s3.GetBucketVersioningInput{
			Bucket:              aws.String(profile.Bucket),
			ExpectedBucketOwner: aws.String(profile.ExpectedOwner),
		},
	)
	if err != nil || versioning == nil ||
		versioning.Status != types.BucketVersioningStatusEnabled {
		return ErrPreflight
	}
	return nil
}
