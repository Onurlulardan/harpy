package documents

import (
	"context"
	"net/http"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) dispatchCompatibility(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	if !onlyHTTPQuery(request.Query) || len(request.Document) != 0 || request.Content != nil {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	contract := "documents_stream_upload_v1"
	if request.Operation == "documents.compat.legacy_sync_archive" {
		contract = "documents_async_archive_v1"
	}
	var result adoptionapp.CompatibilityHit
	err := runtime.command(ctx, request.Access, request.Operation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			if authorizeErr := (applicationAuthorizer{}).Authorize(scoped, scope,
				request.Access.Permission()); authorizeErr != nil {
				return authorizeErr
			}
			hit := adoptionapp.CompatibilityHit{RouteContract: contract,
				ConsumerID: request.Access.PrincipalID().String(),
				BuildID:    runtime.dependencies.BuildID}
			var recordErr error
			if request.Operation == "documents.compat.legacy_sync_archive" {
				result, recordErr = runtime.dependencies.Adoption.ReceiveLegacyArchiveHit(
					scoped, scope, hit)
			} else {
				result, recordErr = runtime.dependencies.Adoption.MarkLegacyUploadHit(
					scoped, scope, hit)
			}
			return recordErr
		})
	data := map[string]any{"contract": result.RouteContract, "hit_count": result.HitCount}
	return documentshttp.Response{Status: http.StatusOK, Data: data},
		mapHTTPApplicationError(err)
}
