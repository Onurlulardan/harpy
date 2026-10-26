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

func (writer *Writer) verifyWritten(
	ctx context.Context,
	evidence WriteEvidence,
) (WriteEvidence, error) {
	output, err := writer.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(writer.bucket), Key: aws.String(evidence.Key),
		VersionId:           aws.String(evidence.VersionID),
		ExpectedBucketOwner: aws.String(writer.owner),
	})
	if err != nil || output == nil || output.Body == nil {
		return WriteEvidence{}, ErrObject
	}
	hash := sha256.New()
	bytes, copyErr := io.Copy(hash, output.Body)
	closeErr := output.Body.Close()
	digest := hex.EncodeToString(hash.Sum(nil))
	if copyErr != nil || closeErr != nil ||
		bytes != evidence.Bytes || digest != evidence.SHA256 {
		return WriteEvidence{}, ErrObject
	}
	retention, err := writer.client.GetObjectRetention(
		ctx, &s3.GetObjectRetentionInput{
			Bucket: aws.String(writer.bucket), Key: aws.String(evidence.Key),
			VersionId:           aws.String(evidence.VersionID),
			ExpectedBucketOwner: aws.String(writer.owner),
		},
	)
	if err != nil || retention == nil || retention.Retention == nil ||
		retention.Retention.Mode != types.ObjectLockRetentionModeCompliance ||
		retention.Retention.RetainUntilDate == nil ||
		retention.Retention.RetainUntilDate.Before(writer.retainUntil) {
		return WriteEvidence{}, ErrObject
	}
	evidence.ReadbackSHA256 = digest
	evidence.ReadbackBytes = bytes
	evidence.RetentionMode = "COMPLIANCE"
	evidence.RetainUntil = retention.Retention.RetainUntilDate.UTC()
	return evidence, nil
}
