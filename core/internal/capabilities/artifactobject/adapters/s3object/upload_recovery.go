package s3object

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

func (store *Store) recoverUpload(ctx context.Context, request artifactobject.Upload,
	putErr error) (artifactobject.UploadResult, error) {
	probe, probeErr := store.api.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &store.bucket, Key: &request.Key, ExpectedBucketOwner: store.owner,
		ChecksumMode: types.ChecksumModeEnabled})
	if probeErr == nil {
		return store.recoveredUpload(request, putErr, probe)
	}
	if provesAbsence(probeErr) && knownNoWrite(putErr) {
		return artifactobject.UploadResult{}, providerFailure(putErr)
	}
	return artifactobject.UploadResult{}, errors.Join(
		artifactobject.ErrOutcomeUnknown, putErr, probeErr)
}

func (store *Store) recoveredUpload(request artifactobject.Upload, putErr error,
	probe *s3.HeadObjectOutput) (artifactobject.UploadResult, error) {
	if probe == nil || probe.VersionId == nil || *probe.VersionId == "" ||
		probe.ContentLength == nil || *probe.ContentLength != request.CiphertextBytes ||
		(probe.DeleteMarker != nil && *probe.DeleteMarker) ||
		store.verifyMetadata(probe.Metadata, probe.ChecksumSHA256,
			request.CiphertextSHA256, request.HeaderSHA256,
			request.CiphertextBytes) != nil {
		return artifactobject.UploadResult{}, errors.Join(
			artifactobject.ErrOutcomeUnknown, putErr)
	}
	requestID := errorRequestID(putErr)
	if requestID == "" {
		var err error
		requestID, err = metadataRequestID(probe.ResultMetadata)
		if err != nil {
			return artifactobject.UploadResult{}, errors.Join(
				artifactobject.ErrOutcomeUnknown, err)
		}
	}
	uploadedAt, err := store.evidenceNow()
	if err != nil {
		return artifactobject.UploadResult{}, err
	}
	return artifactobject.UploadResult{Identity: artifactobject.Identity{
		Key: request.Key, Version: *probe.VersionId}, ProviderRequestID: requestID,
		UploadedAt: uploadedAt}, nil
}
