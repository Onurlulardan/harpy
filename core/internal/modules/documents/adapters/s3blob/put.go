package s3blob

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
)

func (store *Store) Put(ctx context.Context,
	request uploadapp.ObjectPutRequest,
) (uploadapp.ObjectPutResult, error) {
	if store == nil || ctx == nil || ctx.Err() != nil || !validKey(request.Key) ||
		request.Generation == 0 || request.CiphertextBytes < 1 || request.Body == nil {
		return uploadapp.ObjectPutResult{}, uploadapp.ErrInvalid
	}
	contentType := "application/octet-stream"
	output, err := store.api.PutObject(ctx, &s3.PutObjectInput{Bucket: &store.bucket,
		Key: &request.Key, Body: request.Body, ContentLength: &request.CiphertextBytes,
		ContentType: &contentType, ExpectedBucketOwner: store.owner,
		ChecksumAlgorithm: types.ChecksumAlgorithmSha256,
		Metadata:          metadata(request.Generation, request.CiphertextBytes)})
	if err != nil {
		return uploadapp.ObjectPutResult{}, errors.Join(uploadapp.ErrUnavailable, err)
	}
	if output == nil || output.VersionId == nil || *output.VersionId == "" ||
		output.ChecksumSHA256 == nil || *output.ChecksumSHA256 == "" {
		return uploadapp.ObjectPutResult{}, errors.Join(uploadapp.ErrUnavailable, ErrIntegrity)
	}
	return uploadapp.ObjectPutResult{Version: *output.VersionId,
		Generation: request.Generation}, nil
}
