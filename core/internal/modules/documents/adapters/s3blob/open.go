package s3blob

import (
	"context"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
)

func (store *Store) OpenExact(ctx context.Context,
	expected downloadapp.ObjectExpected,
) (io.ReadCloser, error) {
	if store == nil || ctx == nil || ctx.Err() != nil || !validKey(expected.Key) ||
		expected.Version == "" || expected.Generation == 0 ||
		expected.CiphertextBytes < 1 || expected.CiphertextSHA256 == [32]byte{} {
		return nil, downloadapp.ErrInvalid
	}
	output, err := store.api.GetObject(ctx, &s3.GetObjectInput{Bucket: &store.bucket,
		Key: &expected.Key, VersionId: &expected.Version,
		ExpectedBucketOwner: store.owner, ChecksumMode: types.ChecksumModeEnabled})
	if err != nil {
		return nil, errors.Join(downloadapp.ErrUnavailable, err)
	}
	matched := output != nil && output.Body != nil && output.VersionId != nil &&
		*output.VersionId == expected.Version && output.ContentLength != nil &&
		*output.ContentLength == expected.CiphertextBytes &&
		output.ChecksumSHA256 != nil && *output.ChecksumSHA256 == checksum(expected.CiphertextSHA256) &&
		matchingMetadata(output.Metadata, expected.Generation, expected.CiphertextBytes)
	if !matched {
		if output != nil && output.Body != nil {
			_ = output.Body.Close()
		}
		return nil, errors.Join(downloadapp.ErrIntegrity, ErrIntegrity)
	}
	return output.Body, nil
}

var _ downloadapp.ObjectStore = (*Store)(nil)
