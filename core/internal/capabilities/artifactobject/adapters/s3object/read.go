package s3object

import (
	"context"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

func (store *Store) OpenExact(ctx context.Context,
	expected artifactobject.Expected) (io.ReadCloser, error) {
	if err := store.operationError(ctx); err != nil {
		return nil, err
	}
	if !store.validObjectKey(expected.Identity.Key) || expected.Identity.Version == "" ||
		expected.CiphertextBytes < 1 || expected.CiphertextBytes > store.maximumBytes ||
		expected.CiphertextSHA256 == [32]byte{} || expected.HeaderSHA256 == [32]byte{} {
		return nil, ErrObjectInput
	}
	output, err := store.api.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &store.bucket, Key: &expected.Identity.Key,
		VersionId: &expected.Identity.Version, ExpectedBucketOwner: store.owner,
		ChecksumMode: types.ChecksumModeEnabled})
	if err != nil {
		if provesAbsence(err) {
			return nil, errors.Join(artifactobject.ErrIntegrity, err)
		}
		return nil, providerFailure(err)
	}
	if err := store.verifyGetOutput(output, expected); err != nil {
		if output != nil && output.Body != nil {
			err = errors.Join(err, output.Body.Close())
		}
		return nil, errors.Join(artifactobject.ErrIntegrity, err)
	}
	return output.Body, nil
}

func (store *Store) verifyGetOutput(output *s3.GetObjectOutput,
	expected artifactobject.Expected) error {
	if output == nil || output.Body == nil || output.VersionId == nil ||
		*output.VersionId != expected.Identity.Version || output.ContentLength == nil ||
		*output.ContentLength != expected.CiphertextBytes ||
		(output.DeleteMarker != nil && *output.DeleteMarker) {
		return ErrProviderEvidence
	}
	return store.verifyMetadata(output.Metadata, output.ChecksumSHA256,
		expected.CiphertextSHA256, expected.HeaderSHA256, expected.CiphertextBytes)
}
