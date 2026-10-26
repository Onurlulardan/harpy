package s3blob

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
)

func (store *Store) Reconcile(ctx context.Context,
	expected uploadapp.ObjectExpected,
) (uploadapp.ObjectPutResult, error) {
	if store == nil || ctx == nil || ctx.Err() != nil || !validExpected(expected) ||
		expected.Version != "" {
		return uploadapp.ObjectPutResult{}, uploadapp.ErrInvalid
	}
	versions, err := store.listExactVersions(ctx, expected.Key)
	if err != nil {
		return uploadapp.ObjectPutResult{}, errors.Join(uploadapp.ErrUnavailable, err)
	}
	matches := make([]string, 0, 1)
	for _, version := range versions {
		candidate := expected
		candidate.Version = version
		matched, headErr := store.headMatches(ctx, candidate)
		if headErr != nil {
			return uploadapp.ObjectPutResult{}, errors.Join(uploadapp.ErrUnavailable, headErr)
		}
		if matched {
			matches = append(matches, version)
		}
	}
	if len(matches) != 1 {
		return uploadapp.ObjectPutResult{}, errors.Join(uploadapp.ErrUnavailable, ErrAmbiguous)
	}
	return uploadapp.ObjectPutResult{Version: matches[0],
		Generation: expected.Generation}, nil
}

func (store *Store) listExactVersions(ctx context.Context, key string) ([]string, error) {
	result := make([]string, 0, 2)
	var keyMarker, versionMarker *string
	maximum := int32(1000)
	for {
		output, err := store.api.ListObjectVersions(ctx, &s3.ListObjectVersionsInput{
			Bucket: &store.bucket, Prefix: &key, KeyMarker: keyMarker,
			VersionIdMarker: versionMarker, MaxKeys: &maximum,
			ExpectedBucketOwner: store.owner})
		if err != nil || output == nil {
			return nil, errors.Join(ErrAmbiguous, err)
		}
		for _, version := range output.Versions {
			if version.Key != nil && *version.Key == key && version.VersionId != nil &&
				*version.VersionId != "" {
				result = append(result, *version.VersionId)
			}
		}
		if output.IsTruncated == nil || !*output.IsTruncated {
			return result, nil
		}
		if output.NextKeyMarker == nil || output.NextVersionIdMarker == nil {
			return nil, ErrAmbiguous
		}
		keyMarker, versionMarker = output.NextKeyMarker, output.NextVersionIdMarker
	}
}
