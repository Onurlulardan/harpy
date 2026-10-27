package recoverys3

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type RehydrateAPI interface {
	ReadAPI
	PutObject(
		context.Context, *s3.PutObjectInput, ...func(*s3.Options),
	) (*s3.PutObjectOutput, error)
	ListObjectVersions(
		context.Context, *s3.ListObjectVersionsInput, ...func(*s3.Options),
	) (*s3.ListObjectVersionsOutput, error)
}

type RehydrateWriter struct {
	client      RehydrateAPI
	bucket      string
	owner       string
	operationID string
}

func NewRehydrateWriter(
	client RehydrateAPI,
	bucket, owner, operationID string,
) *RehydrateWriter {
	if client == nil || !validBucket(bucket) || owner == "" ||
		!validOperationID(operationID) {
		return nil
	}
	return &RehydrateWriter{
		client: client, bucket: bucket, owner: owner, operationID: operationID,
	}
}

func (writer *RehydrateWriter) PutVerified(
	ctx context.Context,
	request WriteRequest,
) (WriteEvidence, error) {
	if !writer.validRequest(ctx, request) {
		return WriteEvidence{}, ErrObject
	}
	if existing, found, err := writer.findExisting(ctx, request); err != nil {
		return WriteEvidence{}, err
	} else if found {
		return existing, nil
	}
	hash, counter := sha256.New(), &countWriter{}
	output, err := writer.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(writer.bucket), Key: aws.String(request.Key),
		Body:                io.TeeReader(request.Body, io.MultiWriter(hash, counter)),
		ContentLength:       aws.Int64(request.ContentLength),
		ExpectedBucketOwner: aws.String(writer.owner),
		ChecksumSHA256:      nil,
		Metadata: map[string]string{
			"harpy-restore-operation-id": writer.operationID,
			"harpy-plaintext-sha256":     request.ExpectedSHA256,
		},
	})
	uploadedDigest := hex.EncodeToString(hash.Sum(nil))
	if err != nil {
		if existing, found, reconcileErr := writer.findExisting(
			ctx, request,
		); reconcileErr == nil && found {
			return existing, nil
		}
		return WriteEvidence{}, ErrObject
	}
	if output == nil || output.VersionId == nil ||
		!validVersion(*output.VersionId) ||
		counter.bytes != request.ExpectedBytes ||
		uploadedDigest != request.ExpectedSHA256 {
		return WriteEvidence{}, ErrObject
	}
	return writer.verifyExact(ctx, ObjectIdentity{
		Key: request.Key, VersionID: *output.VersionId,
		SHA256: request.ExpectedSHA256, Bytes: request.ExpectedBytes,
	})
}

func (writer *RehydrateWriter) validRequest(
	ctx context.Context,
	request WriteRequest,
) bool {
	return writer != nil && ctx != nil && ctx.Err() == nil &&
		request.Body != nil && validObjectKey(request.Key) &&
		request.ContentLength > 0 &&
		request.ContentLength == request.ExpectedBytes &&
		objectDigestPattern.MatchString(request.ExpectedSHA256)
}
