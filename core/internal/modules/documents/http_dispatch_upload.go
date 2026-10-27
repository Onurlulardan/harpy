package documents

import (
	"context"
	"net/http"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type httpCreateUploadRequest struct {
	FolderID                 string   `json:"folder_id"`
	DocumentID               string   `json:"document_id"`
	ExpectedCurrentVersionID string   `json:"expected_current_version_id"`
	OriginalName             string   `json:"original_name"`
	DeclaredMIME             string   `json:"declared_mime"`
	ExpectedBytes            int64    `json:"expected_bytes"`
	ExpectedSHA256           string   `json:"expected_sha256"`
	Description              string   `json:"description"`
	Tags                     []string `json:"tags"`
}

func (runtime *HTTPRuntime) dispatchUpload(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	switch request.Operation {
	case uploadapp.OperationCreate:
		return runtime.createHTTPUpload(ctx, request)
	case uploadapp.OperationGet:
		return runtime.getHTTPUpload(ctx, request)
	case uploadapp.OperationPut:
		return runtime.putHTTPUpload(ctx, request)
	case uploadapp.OperationCancel:
		return runtime.cancelHTTPUpload(ctx, request)
	default:
		return documentshttp.Response{}, documentshttp.ErrForbidden
	}
}

func (runtime *HTTPRuntime) createHTTPUpload(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	input, err := decodeHTTPRequest[httpCreateUploadRequest](request.Document)
	if err != nil || !onlyHTTPQuery(request.Query) {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	folderID, folderErr := parseHTTPID(input.FolderID, false)
	documentID, documentErr := parseHTTPID(input.DocumentID, true)
	currentID, currentErr := parseHTTPID(input.ExpectedCurrentVersionID, true)
	digest, digestErr := parseHTTPDigest(input.ExpectedSHA256)
	if folderErr != nil || documentErr != nil || currentErr != nil || digestErr != nil {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	var result uploadapp.Snapshot
	err = runtime.command(ctx, request.Access, request.Operation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var createErr error
			result, createErr = runtime.dependencies.Uploads.Create(scoped, scope,
				uploadapp.CreateRequest{FolderID: folderID, DocumentID: documentID,
					ExpectedCurrentVersionID: currentID, OriginalName: input.OriginalName,
					DeclaredMIME: input.DeclaredMIME, ExpectedBytes: input.ExpectedBytes,
					ExpectedContentSHA256: digest, Description: input.Description,
					Tags: input.Tags, IdempotencyKey: request.IdempotencyKey})
			return createErr
		})
	return documentshttp.Response{Status: http.StatusCreated, Data: httpUpload(result),
			Location: "/api/document-uploads/" + result.ID.String()},
		mapHTTPApplicationError(err)
}

func (runtime *HTTPRuntime) getHTTPUpload(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	uploadID, err := parseHTTPID(request.ResourceID, false)
	if err != nil || !onlyHTTPQuery(request.Query) {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	var result uploadapp.Snapshot
	err = runtime.snapshot(ctx, request.Access, request.Operation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var getErr error
			result, getErr = runtime.dependencies.Uploads.Get(scoped, scope,
				uploadapp.GetRequest{UploadID: uploadID})
			return getErr
		})
	return documentshttp.Response{Status: http.StatusOK, Data: httpUpload(result)},
		mapHTTPApplicationError(err)
}
