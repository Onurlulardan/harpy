package s3blob

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	uploadreconcileapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/uploadreconcile"
)

func (store *Store) FindExactVersion(ctx context.Context, key string,
	generation uint64, ciphertextBytes int64,
) (string, bool, error) {
	if store == nil || ctx == nil || ctx.Err() != nil || !validKey(key) ||
		generation == 0 || ciphertextBytes < 1 {
		return "", false, uploadreconcileapp.ErrInvalid
	}
	versions, err := store.listExactVersions(ctx, key)
	if err != nil {
		return "", false, errors.Join(uploadreconcileapp.ErrUnavailable, err)
	}
	matches := make([]string, 0, 1)
	for _, version := range versions {
		output, headErr := store.api.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: &store.bucket, Key: &key, VersionId: &version,
			ExpectedBucketOwner: store.owner})
		if headErr != nil || output == nil {
			return "", false, errors.Join(uploadreconcileapp.ErrUnavailable, headErr)
		}
		if output.VersionId != nil && *output.VersionId == version &&
			output.ContentLength != nil && *output.ContentLength == ciphertextBytes &&
			matchingMetadata(output.Metadata, generation, ciphertextBytes) {
			matches = append(matches, version)
		}
	}
	if len(matches) > 1 {
		return "", false, errors.Join(uploadreconcileapp.ErrUnavailable, ErrAmbiguous)
	}
	if len(matches) == 0 {
		return "", false, nil
	}
	return matches[0], true, nil
}

var _ uploadreconcileapp.Inventory = (*Store)(nil)
