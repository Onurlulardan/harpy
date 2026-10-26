package systemtransferartifact

import (
	"context"
	"errors"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

func publishDraft(ctx context.Context, objects artifactobject.Uploader,
	draft sealedDraft) (result artifactobject.UploadResult, err error) {
	if ctx == nil || ctx.Err() != nil || nilDependency(objects) || draft.lease == nil ||
		draft.objectKey == "" || draft.ciphertext.Validate() != nil {
		return result, errInvalid
	}
	reader, err := draft.lease.OpenReader()
	if err != nil {
		return result, errors.Join(errUnavailable, err)
	}
	seekable, ok := reader.(interface {
		Read([]byte) (int, error)
		Seek(int64, int) (int64, error)
		Close() error
	})
	if !ok {
		_ = reader.Close()
		return result, errInvalid
	}
	result, uploadErr := objects.Upload(ctx, artifactobject.Upload{
		Key: draft.objectKey, Body: seekable,
		CiphertextBytes:  int64(draft.ciphertext.CiphertextBytes),
		CiphertextSHA256: draft.ciphertext.CiphertextSHA256,
		HeaderSHA256:     draft.ciphertext.HeaderSHA256})
	closeErr := reader.Close()
	if uploadErr != nil || closeErr != nil {
		return result, errors.Join(uploadErr, closeErr)
	}
	if result.Identity.Key != draft.objectKey || result.Identity.Version == "" ||
		result.ProviderRequestID == "" || result.UploadedAt.IsZero() {
		return result, errOutcomeUnknown
	}
	return result, nil
}
