package s3object

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type API interface {
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	HeadObject(context.Context, *s3.HeadObjectInput, ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(context.Context, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	ListObjectVersions(context.Context, *s3.ListObjectVersionsInput,
		...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error)
	GetBucketVersioning(context.Context, *s3.GetBucketVersioningInput, ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error)
	GetBucketAcl(context.Context, *s3.GetBucketAclInput, ...func(*s3.Options)) (*s3.GetBucketAclOutput, error)
}

type Clock interface{ Now() time.Time }

type Config struct {
	Bucket                 string
	ExpectedBucketOwner    string
	ObjectPrefix           string
	Contract               string
	MaximumCiphertextBytes int64
}

type Store struct {
	api                      API
	clock                    Clock
	bucket, prefix, contract string
	owner                    *string
	maximumBytes             int64
}

func NewVerified(ctx context.Context, client API, config Config, clock Clock) (*Store, error) {
	if ctx == nil || ctx.Err() != nil || nilDependency(client) || nilDependency(clock) ||
		!validConfig(config) {
		return nil, ErrConfiguration
	}
	store := &Store{api: client, clock: clock, bucket: config.Bucket,
		prefix: config.ObjectPrefix, contract: config.Contract,
		maximumBytes: config.MaximumCiphertextBytes}
	if config.ExpectedBucketOwner != "" {
		store.owner = &config.ExpectedBucketOwner
	}
	versioning, err := client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
		Bucket: &store.bucket, ExpectedBucketOwner: store.owner})
	if err != nil {
		return nil, fmt.Errorf("verify artifact bucket versioning: %w", err)
	}
	if versioning == nil || versioning.Status != types.BucketVersioningStatusEnabled {
		return nil, ErrVersioningRequired
	}
	acl, err := client.GetBucketAcl(ctx, &s3.GetBucketAclInput{
		Bucket: &store.bucket, ExpectedBucketOwner: store.owner})
	if err != nil {
		return nil, fmt.Errorf("verify artifact bucket ACL: %w", err)
	}
	if !privateBucketACL(acl, config.ExpectedBucketOwner) {
		return nil, ErrPrivateACLRequired
	}
	return store, nil
}

func validConfig(config Config) bool {
	return config.Bucket != "" && strings.TrimSpace(config.Bucket) == config.Bucket &&
		strings.TrimSpace(config.ExpectedBucketOwner) == config.ExpectedBucketOwner &&
		validToken(config.ObjectPrefix, 63, true) && validToken(config.Contract, 96, false) &&
		config.MaximumCiphertextBytes >= 1 && config.MaximumCiphertextBytes <= 2147483648
}
