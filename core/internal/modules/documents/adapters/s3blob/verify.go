package s3blob

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
)

func (store *Store) Verify(ctx context.Context, expected uploadapp.ObjectExpected) error {
	if store == nil || ctx == nil || ctx.Err() != nil || !validExpected(expected) ||
		expected.Version == "" {
		return uploadapp.ErrInvalid
	}
	matched, err := store.headMatches(ctx, expected)
	if err != nil {
		return errors.Join(uploadapp.ErrUnavailable, err)
	}
	if !matched {
		return errors.Join(uploadapp.ErrUnavailable, ErrIntegrity)
	}
	return nil
}

func (store *Store) headMatches(ctx context.Context,
	expected uploadapp.ObjectExpected,
) (bool, error) {
	output, err := store.api.HeadObject(ctx, &s3.HeadObjectInput{Bucket: &store.bucket,
		Key: &expected.Key, VersionId: &expected.Version, ExpectedBucketOwner: store.owner,
		ChecksumMode: types.ChecksumModeEnabled})
	if err != nil {
		return false, err
	}
	return output != nil && output.VersionId != nil &&
		*output.VersionId == expected.Version && output.ContentLength != nil &&
		*output.ContentLength == expected.CiphertextBytes &&
		output.ChecksumSHA256 != nil && *output.ChecksumSHA256 == checksum(expected.CiphertextSHA256) &&
		matchingMetadata(output.Metadata, expected.Generation, expected.CiphertextBytes), nil
}

func validExpected(expected uploadapp.ObjectExpected) bool {
	return validKey(expected.Key) && expected.Generation > 0 && expected.CiphertextBytes > 0 &&
		expected.CiphertextSHA256 != [32]byte{}
}
