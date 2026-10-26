package recoverys3

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoverystorage"
)

var ErrPreflight = errors.New("recovery S3 preflight failed")

type API interface {
	GetBucketAcl(context.Context, *s3.GetBucketAclInput, ...func(*s3.Options)) (*s3.GetBucketAclOutput, error)
	GetBucketVersioning(context.Context, *s3.GetBucketVersioningInput, ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error)
	GetObjectLockConfiguration(context.Context, *s3.GetObjectLockConfigurationInput,
		...func(*s3.Options)) (*s3.GetObjectLockConfigurationOutput, error)
}

type Evidence struct {
	Bucket        string
	Owner         string
	Versioned     bool
	ObjectLocked  bool
	RetentionMode string
	RetentionDays int
}

func Preflight(
	ctx context.Context,
	client API,
	manifest recoverystorage.Manifest,
) (Evidence, error) {
	if ctx == nil || ctx.Err() != nil || client == nil || !manifest.Valid() {
		return Evidence{}, ErrPreflight
	}
	bucket, owner := manifest.Bucket, manifest.ExpectedOwner
	acl, err := client.GetBucketAcl(ctx,
		&s3.GetBucketAclInput{Bucket: &bucket, ExpectedBucketOwner: &owner})
	if err != nil || !privateACL(acl, owner) {
		return Evidence{}, ErrPreflight
	}
	versioning, err := client.GetBucketVersioning(ctx,
		&s3.GetBucketVersioningInput{Bucket: &bucket, ExpectedBucketOwner: &owner})
	if err != nil || versioning == nil || versioning.Status != types.BucketVersioningStatusEnabled {
		return Evidence{}, ErrPreflight
	}
	lock, err := client.GetObjectLockConfiguration(ctx,
		&s3.GetObjectLockConfigurationInput{Bucket: &bucket, ExpectedBucketOwner: &owner})
	if err != nil || !validLock(lock, manifest.ObjectLock.RetentionDays) {
		return Evidence{}, ErrPreflight
	}
	return Evidence{Bucket: bucket, Owner: owner, Versioned: true, ObjectLocked: true,
		RetentionMode: "COMPLIANCE", RetentionDays: manifest.ObjectLock.RetentionDays}, nil
}

func validLock(output *s3.GetObjectLockConfigurationOutput, days int) bool {
	if output == nil || output.ObjectLockConfiguration == nil ||
		output.ObjectLockConfiguration.ObjectLockEnabled != types.ObjectLockEnabledEnabled ||
		output.ObjectLockConfiguration.Rule == nil ||
		output.ObjectLockConfiguration.Rule.DefaultRetention == nil {
		return false
	}
	retention := output.ObjectLockConfiguration.Rule.DefaultRetention
	return retention.Mode == types.ObjectLockRetentionModeCompliance &&
		retention.Days != nil && int(*retention.Days) == days && retention.Years == nil
}
