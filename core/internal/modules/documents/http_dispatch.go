package documents

import (
	"context"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

func (runtime *HTTPRuntime) Dispatch(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	if runtime == nil || ctx == nil || ctx.Err() != nil ||
		!request.Access.ValidFor(request.Access.Permission(), request.Operation) {
		return documentshttp.Response{}, documentshttp.ErrForbidden
	}
	actorID, actorErr := foundation.ParseID(request.Access.PrincipalID().String())
	tenantID, tenantErr := foundation.ParseID(request.Access.TenantID().String())
	evidence, err := foundation.NewAuthorityEvidence(actorID, tenantID,
		request.Access.TokenVersion(), request.Access.AuthorizationVersion(),
		request.Access.AuthorizationClock(), request.Access.Permission(), request.Operation,
		request.Access.StepUpAuthenticatedAt())
	if actorErr != nil || tenantErr != nil || err != nil {
		return documentshttp.Response{}, documentshttp.ErrUnavailable
	}
	ctx, err = foundation.WithAuthorityEvidence(ctx, evidence)
	if err != nil {
		return documentshttp.Response{}, documentshttp.ErrUnavailable
	}
	switch request.Operation {
	case "documents.folder.create", "documents.folder.get", "documents.folder.list",
		"documents.folder.update", "documents.folder.move", "documents.folder.delete",
		"documents.folder.restore", "documents.folder.contents":
		return runtime.dispatchFolder(ctx, request)
	case "documents.item.get", "documents.item.list", "documents.item.batch",
		"documents.item.update", "documents.item.move", "documents.item.delete",
		"documents.item.restore", "documents.item.purge", "documents.item.cancel_purge",
		"documents.item.versions.list", "documents.item.version.latest":
		return runtime.dispatchItem(ctx, request)
	case "documents.item.copy":
		return runtime.dispatchCopy(ctx, request)
	case "documents.item.download", "documents.item.version.download":
		return runtime.dispatchDownload(ctx, request)
	case "documents.upload.create", "documents.upload.get", "documents.upload.put",
		"documents.upload.cancel":
		return runtime.dispatchUpload(ctx, request)
	case "documents.archive.create", "documents.archive.get",
		"documents.archive.download":
		return runtime.dispatchArchive(ctx, request)
	case "documents.compat.legacy_sync_archive",
		"documents.compat.legacy_multipart_upload":
		return runtime.dispatchCompatibility(ctx, request)
	default:
		return documentshttp.Response{}, documentshttp.ErrForbidden
	}
}
