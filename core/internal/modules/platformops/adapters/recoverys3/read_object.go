package recoverys3

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (reader *Reader) DownloadVerified(
	ctx context.Context,
	identity ObjectIdentity,
	target io.Writer,
) error {
	if reader == nil || ctx == nil || ctx.Err() != nil || target == nil ||
		!validObjectKey(identity.Key) || !validVersion(identity.VersionID) ||
		!validExpected(identity.SHA256, identity.Bytes) {
		return ErrObject
	}
	output, err := reader.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(reader.bucket), Key: aws.String(identity.Key),
		VersionId:           aws.String(identity.VersionID),
		ExpectedBucketOwner: aws.String(reader.owner),
	})
	if err != nil || output == nil || output.Body == nil {
		return ErrObject
	}
	defer output.Body.Close()
	hash := sha256.New()
	count, copyErr := io.Copy(io.MultiWriter(target, hash), output.Body)
	if copyErr != nil || count != identity.Bytes ||
		hex.EncodeToString(hash.Sum(nil)) != identity.SHA256 {
		return ErrObject
	}
	return nil
}

func (reader *Reader) OpenExact(
	ctx context.Context,
	key string,
	versionID string,
) (io.ReadCloser, error) {
	if reader == nil || ctx == nil || ctx.Err() != nil ||
		!validObjectKey(key) || !validVersion(versionID) {
		return nil, ErrObject
	}
	output, err := reader.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(reader.bucket), Key: aws.String(key),
		VersionId:           aws.String(versionID),
		ExpectedBucketOwner: aws.String(reader.owner),
	})
	if err != nil || output == nil || output.Body == nil {
		return nil, ErrObject
	}
	return output.Body, nil
}
