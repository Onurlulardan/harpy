package recoverys3

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

var ErrObject = errors.New("recovery S3 object operation failed")

type ReadAPI interface {
	GetObject(
		context.Context,
		*s3.GetObjectInput,
		...func(*s3.Options),
	) (*s3.GetObjectOutput, error)
}

type WriteAPI interface {
	ReadAPI
	PutObject(
		context.Context,
		*s3.PutObjectInput,
		...func(*s3.Options),
	) (*s3.PutObjectOutput, error)
	GetObjectRetention(
		context.Context,
		*s3.GetObjectRetentionInput,
		...func(*s3.Options),
	) (*s3.GetObjectRetentionOutput, error)
}

type Reader struct {
	client ReadAPI
	bucket string
	owner  string
}

type Writer struct {
	client      WriteAPI
	bucket      string
	owner       string
	retainUntil time.Time
}

type ObjectIdentity = ports.ObjectIdentity
type WriteRequest = ports.WriteRequest
type WriteEvidence = ports.WriteEvidence

func NewReader(client ReadAPI, bucket, owner string) *Reader {
	if client == nil || !validBucket(bucket) || owner == "" {
		return nil
	}
	return &Reader{client: client, bucket: bucket, owner: owner}
}

func NewWriter(
	client WriteAPI,
	bucket string,
	owner string,
	retentionDays int,
	boundNow time.Time,
) *Writer {
	if client == nil || !validBucket(bucket) || owner == "" ||
		retentionDays < 1 || retentionDays > 36500 || boundNow.IsZero() {
		return nil
	}
	return &Writer{client: client, bucket: bucket, owner: owner,
		retainUntil: boundNow.UTC().Add(
			time.Duration(retentionDays) * 24 * time.Hour,
		)}
}
