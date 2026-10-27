package s3object

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

func (store *Store) Upload(ctx context.Context,
	request artifactobject.Upload) (artifactobject.UploadResult, error) {
	if err := store.operationError(ctx); err != nil {
		return artifactobject.UploadResult{}, err
	}
	if request.Body == nil || !store.validObjectKey(request.Key) ||
		request.CiphertextBytes < 1 || request.CiphertextBytes > store.maximumBytes ||
		request.CiphertextSHA256 == [32]byte{} || request.HeaderSHA256 == [32]byte{} {
		return artifactobject.UploadResult{}, ErrObjectInput
	}
	if err := verifySeekableBody(request.Body, request.CiphertextBytes); err != nil {
		return artifactobject.UploadResult{}, err
	}
	checksum := checksumBase64(request.CiphertextSHA256)
	contentType, ifNoneMatch := "application/octet-stream", "*"
	output, err := store.api.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &store.bucket, Key: &request.Key, Body: request.Body,
		ContentLength: &request.CiphertextBytes, ContentType: &contentType,
		ChecksumAlgorithm: types.ChecksumAlgorithmSha256, ChecksumSHA256: &checksum,
		ExpectedBucketOwner: store.owner, IfNoneMatch: &ifNoneMatch,
		Metadata: store.objectMetadata(request.CiphertextSHA256,
			request.HeaderSHA256, request.CiphertextBytes)})
	if err != nil {
		return store.recoverUpload(ctx, request, err)
	}
	if output == nil || output.VersionId == nil || *output.VersionId == "" ||
		output.ChecksumSHA256 == nil || *output.ChecksumSHA256 != checksum {
		return store.recoverUpload(ctx, request, ErrProviderEvidence)
	}
	requestID, evidenceErr := metadataRequestID(output.ResultMetadata)
	if evidenceErr != nil {
		return store.recoverUpload(ctx, request, evidenceErr)
	}
	uploadedAt, evidenceErr := store.evidenceNow()
	if evidenceErr != nil {
		return artifactobject.UploadResult{}, evidenceErr
	}
	return artifactobject.UploadResult{Identity: artifactobject.Identity{
		Key: request.Key, Version: *output.VersionId}, ProviderRequestID: requestID,
		UploadedAt: uploadedAt}, nil
}

func verifySeekableBody(body io.ReadSeeker, expected int64) error {
	size, err := body.Seek(0, io.SeekEnd)
	if err != nil || size != expected {
		return fmt.Errorf("%w: ciphertext length", ErrObjectInput)
	}
	position, err := body.Seek(0, io.SeekStart)
	if err != nil || position != 0 {
		return fmt.Errorf("%w: ciphertext position", ErrObjectInput)
	}
	return nil
}

func providerFailure(err error) error {
	return errors.Join(artifactobject.ErrUnavailable, err)
}
