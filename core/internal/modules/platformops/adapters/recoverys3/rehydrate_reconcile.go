package recoverys3

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (writer *RehydrateWriter) findExisting(
	ctx context.Context,
	request WriteRequest,
) (WriteEvidence, bool, error) {
	output, err := writer.client.ListObjectVersions(
		ctx, &s3.ListObjectVersionsInput{
			Bucket: aws.String(writer.bucket), Prefix: aws.String(request.Key),
			ExpectedBucketOwner: aws.String(writer.owner),
			MaxKeys:             aws.Int32(1000),
		},
	)
	if err != nil || output == nil || output.IsTruncated != nil && *output.IsTruncated {
		return WriteEvidence{}, false, ErrObject
	}
	for _, version := range output.Versions {
		if version.Key == nil || *version.Key != request.Key ||
			version.VersionId == nil || !validVersion(*version.VersionId) {
			continue
		}
		identity := ObjectIdentity{
			Key: request.Key, VersionID: *version.VersionId,
			SHA256: request.ExpectedSHA256, Bytes: request.ExpectedBytes,
		}
		evidence, matches, verifyErr := writer.matchesExact(ctx, identity)
		if verifyErr != nil {
			return WriteEvidence{}, false, verifyErr
		}
		if matches {
			return evidence, true, nil
		}
	}
	return WriteEvidence{}, false, nil
}

func (writer *RehydrateWriter) matchesExact(
	ctx context.Context,
	identity ObjectIdentity,
) (WriteEvidence, bool, error) {
	output, err := writer.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(writer.bucket), Key: aws.String(identity.Key),
		VersionId:           aws.String(identity.VersionID),
		ExpectedBucketOwner: aws.String(writer.owner),
	})
	if err != nil || output == nil || output.Body == nil {
		return WriteEvidence{}, false, ErrObject
	}
	defer output.Body.Close()
	if output.Metadata["harpy-restore-operation-id"] != writer.operationID {
		return WriteEvidence{}, false, nil
	}
	hash := sha256.New()
	count, copyErr := io.Copy(hash, output.Body)
	digest := hex.EncodeToString(hash.Sum(nil))
	if copyErr != nil {
		return WriteEvidence{}, false, ErrObject
	}
	if output.Metadata["harpy-plaintext-sha256"] != identity.SHA256 ||
		count != identity.Bytes || digest != identity.SHA256 {
		return WriteEvidence{}, false, ErrObject
	}
	return rehydrateEvidence(identity), true, nil
}

func (writer *RehydrateWriter) verifyExact(
	ctx context.Context,
	identity ObjectIdentity,
) (WriteEvidence, error) {
	evidence, matches, err := writer.matchesExact(ctx, identity)
	if err != nil || !matches {
		return WriteEvidence{}, ErrObject
	}
	return evidence, nil
}

func rehydrateEvidence(identity ObjectIdentity) WriteEvidence {
	return WriteEvidence{
		ObjectIdentity: identity, UploadSHA256: identity.SHA256,
		UploadedBytes: identity.Bytes, ReadbackSHA256: identity.SHA256,
		ReadbackBytes: identity.Bytes,
	}
}
