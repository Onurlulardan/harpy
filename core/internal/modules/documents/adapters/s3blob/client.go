package s3blob

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
)

var (
	ErrConfiguration = errors.New("documents S3 configuration is invalid")
	ErrIntegrity     = errors.New("documents S3 object evidence mismatches")
	ErrAmbiguous     = errors.New("documents S3 outcome is ambiguous")
)

type API interface {
	PutObject(context.Context, *s3.PutObjectInput,
		...func(*s3.Options)) (*s3.PutObjectOutput, error)
	HeadObject(context.Context, *s3.HeadObjectInput,
		...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	GetObject(context.Context, *s3.GetObjectInput,
		...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(context.Context, *s3.DeleteObjectInput,
		...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	ListObjectVersions(context.Context, *s3.ListObjectVersionsInput,
		...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error)
	GetBucketVersioning(context.Context, *s3.GetBucketVersioningInput,
		...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error)
	GetBucketAcl(context.Context, *s3.GetBucketAclInput,
		...func(*s3.Options)) (*s3.GetBucketAclOutput, error)
}

type Config struct {
	Bucket              string
	ExpectedBucketOwner string
}

type Clock interface{ Now() time.Time }

type Store struct {
	api    API
	bucket string
	owner  *string
	clock  Clock
}

func NewVerified(ctx context.Context, api API, config Config, clock Clock) (*Store, error) {
	if ctx == nil || ctx.Err() != nil || api == nil || clock == nil || config.Bucket == "" ||
		strings.TrimSpace(config.Bucket) != config.Bucket ||
		strings.TrimSpace(config.ExpectedBucketOwner) != config.ExpectedBucketOwner {
		return nil, ErrConfiguration
	}
	store := &Store{api: api, bucket: config.Bucket, clock: clock}
	if config.ExpectedBucketOwner != "" {
		store.owner = &config.ExpectedBucketOwner
	}
	versioning, err := api.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
		Bucket: &store.bucket, ExpectedBucketOwner: store.owner})
	if err != nil || versioning == nil ||
		versioning.Status != types.BucketVersioningStatusEnabled {
		return nil, errors.Join(ErrConfiguration, err)
	}
	acl, err := api.GetBucketAcl(ctx, &s3.GetBucketAclInput{
		Bucket: &store.bucket, ExpectedBucketOwner: store.owner})
	if err != nil || !privateACL(acl, config.ExpectedBucketOwner) {
		return nil, errors.Join(ErrConfiguration, err)
	}
	return store, nil
}

func privateACL(output *s3.GetBucketAclOutput, expectedOwner string) bool {
	if output == nil || output.Owner == nil || output.Owner.ID == nil ||
		*output.Owner.ID == "" || expectedOwner != "" && *output.Owner.ID != expectedOwner {
		return false
	}
	owner := *output.Owner.ID
	for _, grant := range output.Grants {
		if grant.Grantee == nil || grant.Grantee.Type != types.TypeCanonicalUser ||
			grant.Grantee.ID == nil || *grant.Grantee.ID != owner ||
			grant.Grantee.URI != nil || grant.Grantee.EmailAddress != nil {
			return false
		}
	}
	return true
}

var _ uploadapp.ObjectStore = (*Store)(nil)
