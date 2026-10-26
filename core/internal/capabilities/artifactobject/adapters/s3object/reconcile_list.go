package s3object

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

const reconciliationListLimit int32 = 3

type reconciliationSnapshot struct {
	present      bool
	version      string
	bytes        int64
	lastModified time.Time
	requestID    string
}

func (store *Store) listReconciliationSnapshot(
	ctx context.Context,
	key string,
) (reconciliationSnapshot, error) {
	output, err := store.api.ListObjectVersions(ctx, &s3.ListObjectVersionsInput{
		Bucket: &store.bucket, Prefix: &key, MaxKeys: int32Pointer(reconciliationListLimit),
		ExpectedBucketOwner: store.owner,
	})
	if err != nil {
		return reconciliationSnapshot{}, providerFailure(err)
	}
	if output == nil || output.IsTruncated == nil || *output.IsTruncated ||
		len(output.CommonPrefixes) != 0 || len(output.DeleteMarkers) != 0 ||
		len(output.Versions) > 1 {
		return reconciliationSnapshot{}, errors.Join(
			artifactobject.ErrOutcomeUnknown, ErrProviderEvidence)
	}
	requestID, err := metadataRequestID(output.ResultMetadata)
	if err != nil {
		return reconciliationSnapshot{}, errors.Join(
			artifactobject.ErrOutcomeUnknown, err)
	}
	if len(output.Versions) == 0 {
		return reconciliationSnapshot{requestID: requestID}, nil
	}
	version := output.Versions[0]
	if version.Key == nil || *version.Key != key || version.VersionId == nil ||
		*version.VersionId == "" || version.IsLatest == nil || !*version.IsLatest ||
		version.Size == nil || *version.Size < 1 || version.LastModified == nil ||
		!validProviderTime(*version.LastModified) {
		return reconciliationSnapshot{}, errors.Join(
			artifactobject.ErrOutcomeUnknown, ErrProviderEvidence)
	}
	return reconciliationSnapshot{present: true, version: *version.VersionId,
		bytes: *version.Size, lastModified: version.LastModified.UTC(),
		requestID: requestID}, nil
}

func (snapshot reconciliationSnapshot) sameObjectState(other reconciliationSnapshot) bool {
	return snapshot.present == other.present && snapshot.version == other.version &&
		snapshot.bytes == other.bytes && snapshot.lastModified.Equal(other.lastModified)
}

func validProviderTime(value time.Time) bool {
	return !value.IsZero() && value.Year() >= 2000
}

func int32Pointer(value int32) *int32 { return &value }
