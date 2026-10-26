package documents

import (
	"context"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
)

type httpUpdateItemRequest struct {
	ExpectedVersion uint64   `json:"expected_version"`
	Description     string   `json:"description"`
	Tags            []string `json:"tags"`
}

type httpMoveItemRequest struct {
	ExpectedVersion uint64 `json:"expected_version"`
	FolderID        string `json:"folder_id"`
}

type httpPurgeItemRequest struct {
	ExpectedVersion uint64 `json:"expected_version"`
	Reason          string `json:"reason"`
}

type httpBatchItemsRequest struct {
	DocumentIDs []string `json:"document_ids"`
}

func (runtime *HTTPRuntime) dispatchItem(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	switch request.Operation {
	case itemapp.OperationGet:
		return runtime.getHTTPItem(ctx, request)
	case itemapp.OperationList:
		return runtime.listHTTPItems(ctx, request)
	case itemapp.OperationBatch:
		return runtime.batchHTTPItems(ctx, request)
	case itemapp.OperationListVersions:
		return runtime.listHTTPVersions(ctx, request)
	case itemapp.OperationLatest:
		return runtime.latestHTTPVersion(ctx, request)
	case itemapp.OperationUpdate, itemapp.OperationMove, itemapp.OperationDelete,
		itemapp.OperationRestore, itemapp.OperationPurge, itemapp.OperationCancelPurge:
		return runtime.mutateHTTPItem(ctx, request)
	default:
		return documentshttp.Response{}, documentshttp.ErrForbidden
	}
}
