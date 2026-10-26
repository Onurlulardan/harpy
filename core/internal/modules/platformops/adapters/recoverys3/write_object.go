package recoverys3

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type countWriter struct{ bytes int64 }

func (writer *countWriter) Write(value []byte) (int, error) {
	writer.bytes += int64(len(value))
	return len(value), nil
}

func (writer *Writer) PutVerified(
	ctx context.Context,
	request WriteRequest,
) (WriteEvidence, error) {
	if writer == nil || ctx == nil || ctx.Err() != nil || request.Body == nil ||
		!validObjectKey(request.Key) ||
		request.ContentLength < -1 || request.ExpectedBytes < 0 ||
		request.ExpectedSHA256 != "" &&
			!objectDigestPattern.MatchString(request.ExpectedSHA256) {
		return WriteEvidence{}, ErrObject
	}
	hash, counter := sha256.New(), &countWriter{}
	input := &s3.PutObjectInput{
		Bucket: aws.String(writer.bucket), Key: aws.String(request.Key),
		Body:                      io.TeeReader(request.Body, io.MultiWriter(hash, counter)),
		ExpectedBucketOwner:       aws.String(writer.owner),
		ObjectLockMode:            types.ObjectLockModeCompliance,
		ObjectLockRetainUntilDate: aws.Time(writer.retainUntil),
	}
	if request.ContentLength >= 0 {
		input.ContentLength = aws.Int64(request.ContentLength)
	}
	output, err := writer.client.PutObject(ctx, input)
	uploadDigest := hex.EncodeToString(hash.Sum(nil))
	if err != nil || output == nil || output.VersionId == nil ||
		!validVersion(*output.VersionId) || counter.bytes < 1 ||
		request.ExpectedBytes > 0 && counter.bytes != request.ExpectedBytes ||
		request.ExpectedSHA256 != "" && uploadDigest != request.ExpectedSHA256 {
		return WriteEvidence{}, ErrObject
	}
	evidence := WriteEvidence{ObjectIdentity: ObjectIdentity{
		Key: request.Key, VersionID: *output.VersionId,
		SHA256: uploadDigest, Bytes: counter.bytes,
	}, UploadSHA256: uploadDigest, UploadedBytes: counter.bytes}
	return writer.verifyWritten(ctx, evidence)
}
