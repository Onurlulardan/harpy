package documents

import (
	"context"
	"net/http"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	archivedownloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archivedownload"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type httpCreateArchiveRequest struct {
	DocumentIDs      []string `json:"document_ids"`
	RetryOfArchiveID string   `json:"retry_of_archive_id"`
}

func (runtime *HTTPRuntime) dispatchArchive(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	switch request.Operation {
	case archiveapp.OperationCreate:
		return runtime.createHTTPArchive(ctx, request)
	case archiveapp.OperationGet:
		return runtime.getHTTPArchive(ctx, request)
	case archivedownloadapp.Operation:
		return runtime.downloadHTTPArchive(ctx, request)
	default:
		return documentshttp.Response{}, documentshttp.ErrForbidden
	}
}

func (runtime *HTTPRuntime) createHTTPArchive(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	input, err := decodeHTTPRequest[httpCreateArchiveRequest](request.Document)
	if err != nil || !onlyHTTPQuery(request.Query) || len(input.DocumentIDs) < 1 ||
		len(input.DocumentIDs) > 10000 {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	ids := make([]foundation.ID, len(input.DocumentIDs))
	for index, raw := range input.DocumentIDs {
		ids[index], err = parseHTTPID(raw, false)
		if err != nil {
			return documentshttp.Response{}, err
		}
	}
	retryID, err := parseHTTPID(input.RetryOfArchiveID, true)
	if err != nil {
		return documentshttp.Response{}, err
	}
	var result archiveapp.Snapshot
	err = runtime.command(ctx, request.Access, request.Operation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var createErr error
			result, createErr = runtime.dependencies.Archives.Create(scoped, scope,
				archiveapp.CreateRequest{ItemIDs: ids, RetryOfArchiveID: retryID,
					IdempotencyKey: request.IdempotencyKey})
			return createErr
		})
	return documentshttp.Response{Status: http.StatusAccepted, Data: httpArchive(result),
			Location: "/api/document-archives/" + result.ID.String()},
		mapHTTPApplicationError(err)
}

func (runtime *HTTPRuntime) getHTTPArchive(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	archiveID, err := parseHTTPID(request.ResourceID, false)
	if err != nil || !onlyHTTPQuery(request.Query) {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	var result archiveapp.Snapshot
	err = runtime.snapshot(ctx, request.Access, request.Operation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var getErr error
			result, getErr = runtime.dependencies.Archives.Get(scoped, scope,
				archiveapp.GetRequest{ArchiveID: archiveID})
			return getErr
		})
	return documentshttp.Response{Status: http.StatusOK, Data: httpArchive(result)},
		mapHTTPApplicationError(err)
}

func (runtime *HTTPRuntime) downloadHTTPArchive(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	archiveID, err := parseHTTPID(request.ResourceID, false)
	if err != nil || !onlyHTTPQuery(request.Query) {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	runner := httpScopeRunner{runtime: runtime, access: request.Access}
	result, err := runtime.dependencies.ArchiveDownloads.Open(ctx, runner,
		archivedownloadapp.Request{ArchiveID: archiveID})
	if err != nil {
		return documentshttp.Response{}, mapHTTPApplicationError(err)
	}
	return documentshttp.Response{Status: http.StatusOK,
		Binary: &documentshttp.Binary{Filename: result.Filename, MIME: "application/zip",
			Bytes: result.Bytes, SHA256: result.SHA256, Content: result.Content}}, nil
}
