package dashboardexportartifact

import (
	"context"
	"errors"
	"io"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

type pending struct {
	runtime *Runtime
	lease   exportapp.EncryptedSpool
	draft   exportapp.ArtifactDraft
}

func (value *pending) Draft() exportapp.ArtifactDraft { return value.draft }

func (value *pending) Upload(ctx context.Context) (exportapp.UploadReceipt, error) {
	if value == nil || value.runtime == nil || value.lease == nil || ctx == nil || ctx.Err() != nil {
		return exportapp.UploadReceipt{}, exportapp.ErrInvalid
	}
	reader, err := value.lease.OpenReader()
	if err != nil {
		return exportapp.UploadReceipt{}, errors.Join(exportapp.ErrUnavailable, err)
	}
	defer func() { _ = reader.Close() }()
	seekable, ok := reader.(io.ReadSeeker)
	if !ok {
		return exportapp.UploadReceipt{}, exportapp.ErrUnavailable
	}
	result, err := value.runtime.objects.Upload(ctx, artifactobject.Upload{
		Key: value.draft.ObjectKey, Body: seekable,
		CiphertextBytes:  value.draft.CiphertextBytes,
		CiphertextSHA256: value.draft.CiphertextSHA256,
		HeaderSHA256:     value.draft.HeaderSHA256})
	if errors.Is(err, artifactobject.ErrOutcomeUnknown) {
		return exportapp.UploadReceipt{}, errors.Join(exportapp.ErrOutcomeUnknown, err)
	}
	if err != nil || result.Identity.Key != value.draft.ObjectKey ||
		result.Identity.Version == "" || result.ProviderRequestID == "" || result.UploadedAt.IsZero() {
		return exportapp.UploadReceipt{}, errors.Join(exportapp.ErrUnavailable, err)
	}
	return exportapp.UploadReceipt{ObjectVersion: result.Identity.Version,
		ProviderRequestID: result.ProviderRequestID,
		UploadedAt:        result.UploadedAt.UTC()}, nil
}

func (value *pending) Discard() error {
	if value == nil || value.lease == nil {
		return nil
	}
	err := value.lease.Discard()
	value.lease = nil
	return err
}
