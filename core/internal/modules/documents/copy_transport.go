package documents

import (
	"context"
	"errors"

	copyapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/copy"
	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
)

type copyTransport struct {
	downloads *downloadapp.Application
	uploads   *uploadapp.Application
}

func (transport copyTransport) Open(ctx context.Context, runner copyapp.ScopeRunner,
	itemID, versionID foundation.ID,
) (copyapp.Source, error) {
	result, err := transport.downloads.OpenForCopy(ctx, runner,
		downloadapp.Request{ItemID: itemID, VersionID: versionID})
	if err != nil {
		return copyapp.Source{}, copyTransportError(err)
	}
	return copyapp.Source{ItemID: result.ItemID, VersionID: result.VersionID,
		OriginalName: result.Filename, MIME: result.MIME, Bytes: result.Bytes,
		SHA256: result.SHA256, Content: result.Content}, nil
}

func (transport copyTransport) CreateAndPut(ctx context.Context,
	runner copyapp.ScopeRunner, request copyapp.DestinationRequest,
) (copyapp.Result, error) {
	created, err := transport.uploads.PrepareForCopy(ctx, runner, uploadapp.CreateRequest{
		FolderID: request.FolderID, OriginalName: request.OriginalName,
		DeclaredMIME: request.MIME, Description: request.Description,
		Tags: request.Tags, ExpectedContentSHA256: request.SHA256,
		ExpectedBytes: request.Bytes, IdempotencyKey: request.IdempotencyKey})
	if err != nil {
		return copyapp.Result{}, copyTransportError(err)
	}
	ready, err := transport.uploads.PersistForCopy(ctx, runner, uploadapp.PutRequest{
		UploadID: created.ID, LeaseOwnerID: request.LeaseOwnerID,
		Content: request.Content})
	if err != nil {
		return copyapp.Result{}, copyTransportError(err)
	}
	return copyapp.Result{ItemID: ready.DocumentID, VersionID: ready.VersionID,
		UploadID: created.ID, Bytes: request.Bytes, SHA256: request.SHA256}, nil
}

func copyTransportError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, uploadapp.ErrInvalid), errors.Is(err, downloadapp.ErrInvalid):
		return copyapp.ErrInvalid
	case errors.Is(err, uploadapp.ErrForbidden), errors.Is(err, downloadapp.ErrForbidden):
		return copyapp.ErrForbidden
	case errors.Is(err, uploadapp.ErrConflict), errors.Is(err, uploadapp.ErrInProgress):
		return copyapp.ErrConflict
	default:
		return copyapp.ErrUnavailable
	}
}

var _ copyapp.SourceReader = copyTransport{}
var _ copyapp.DestinationWriter = copyTransport{}
