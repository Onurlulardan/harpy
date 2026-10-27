package documents

import (
	"context"
	"net/http"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	uploadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/upload"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) putHTTPUpload(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	uploadID, err := parseHTTPID(request.ResourceID, false)
	if err != nil || request.Content == nil || !onlyHTTPQuery(request.Query) {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	leaseID, err := runtime.dependencies.IDs.NewUUIDv7()
	if err != nil || !leaseID.UUIDv7() {
		return documentshttp.Response{}, documentshttp.ErrUnavailable
	}
	runner := httpScopeRunner{runtime: runtime, access: request.Access}
	result, err := runtime.dependencies.Uploads.Put(ctx, runner,
		uploadapp.PutRequest{UploadID: uploadID, LeaseOwnerID: leaseID,
			Content: request.Content})
	if err != nil {
		return documentshttp.Response{}, mapHTTPApplicationError(err)
	}
	data := map[string]any{"upload_id": result.Upload.ID.String(),
		"document_id": result.DocumentID.String(), "version_id": result.VersionID.String()}
	return documentshttp.Response{Status: result.Status, Data: data,
		Location: "/api/documents/" + result.DocumentID.String()}, nil
}

func (runtime *HTTPRuntime) cancelHTTPUpload(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	uploadID, err := parseHTTPID(request.ResourceID, false)
	if err != nil || !onlyHTTPQuery(request.Query) {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	if _, err = decodeHTTPRequest[struct{}](request.Document); err != nil {
		return documentshttp.Response{}, err
	}
	var result uploadapp.Snapshot
	err = runtime.command(ctx, request.Access, request.Operation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var cancelErr error
			result, cancelErr = runtime.dependencies.Uploads.Cancel(scoped, scope,
				uploadapp.CancelRequest{UploadID: uploadID})
			return cancelErr
		})
	return documentshttp.Response{Status: http.StatusOK, Data: httpUpload(result)},
		mapHTTPApplicationError(err)
}
