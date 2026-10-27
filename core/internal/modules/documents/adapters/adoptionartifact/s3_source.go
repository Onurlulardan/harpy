package adoptionartifact

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
)

type S3SourceAPI interface {
	GetObject(context.Context, *s3.GetObjectInput,
		...func(*s3.Options)) (*s3.GetObjectOutput, error)
	HeadObject(context.Context, *s3.HeadObjectInput,
		...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	GetBucketVersioning(context.Context, *s3.GetBucketVersioningInput,
		...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error)
	GetBucketAcl(context.Context, *s3.GetBucketAclInput,
		...func(*s3.Options)) (*s3.GetBucketAclOutput, error)
}

type S3Source struct {
	api    S3SourceAPI
	bucket string
	owner  *string
}

func NewVerifiedS3Source(ctx context.Context, api S3SourceAPI,
	bucket, expectedOwner string,
) (*S3Source, error) {
	if ctx == nil || ctx.Err() != nil || api == nil || !validSourceName(bucket) ||
		strings.TrimSpace(expectedOwner) != expectedOwner {
		return nil, adoptionapp.ErrUnavailable
	}
	source := &S3Source{api: api, bucket: bucket}
	if expectedOwner != "" {
		source.owner = &expectedOwner
	}
	versioning, err := api.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
		Bucket: &source.bucket, ExpectedBucketOwner: source.owner})
	if err != nil || versioning == nil ||
		versioning.Status != types.BucketVersioningStatusEnabled {
		return nil, errors.Join(adoptionapp.ErrUnavailable, err)
	}
	acl, err := api.GetBucketAcl(ctx, &s3.GetBucketAclInput{
		Bucket: &source.bucket, ExpectedBucketOwner: source.owner})
	if err != nil || !sourcePrivateACL(acl, expectedOwner) {
		return nil, errors.Join(adoptionapp.ErrUnavailable, err)
	}
	return source, nil
}

func (source *S3Source) OpenExact(ctx context.Context,
	key, revision string,
) (io.ReadCloser, error) {
	if source == nil || ctx == nil || ctx.Err() != nil ||
		!validSourceName(key) || !validSourceName(revision) {
		return nil, adoptionapp.ErrInvalid
	}
	output, err := source.api.GetObject(ctx, &s3.GetObjectInput{Bucket: &source.bucket,
		Key: &key, VersionId: &revision, ExpectedBucketOwner: source.owner})
	if err != nil {
		return nil, errors.Join(adoptionapp.ErrUnavailable, err)
	}
	if output == nil || output.Body == nil || output.VersionId == nil ||
		*output.VersionId != revision {
		if output != nil && output.Body != nil {
			_ = output.Body.Close()
		}
		return nil, adoptionapp.ErrBlocked
	}
	return output.Body, nil
}

func validSourceName(value string) bool {
	if value == "" || len(value) > 1024 || strings.TrimSpace(value) != value {
		return false
	}
	for _, character := range value {
		if character < 0x20 || character == 0x7f {
			return false
		}
	}
	return true
}

func sourcePrivateACL(output *s3.GetBucketAclOutput, expectedOwner string) bool {
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

var _ SourceStore = (*S3Source)(nil)
