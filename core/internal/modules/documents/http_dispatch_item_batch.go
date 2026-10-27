package documents

import (
	"context"
	"net/http"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) batchHTTPItems(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	input, err := decodeHTTPRequest[httpBatchItemsRequest](request.Document)
	if err != nil || !onlyHTTPQuery(request.Query) || len(input.DocumentIDs) < 1 ||
		len(input.DocumentIDs) > 100 {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	ids := make([]foundation.ID, len(input.DocumentIDs))
	for index, raw := range input.DocumentIDs {
		ids[index], err = parseHTTPID(raw, false)
		if err != nil {
			return documentshttp.Response{}, err
		}
	}
	var snapshots []httpItemView
	err = runtime.snapshot(ctx, request.Access, request.Operation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			items, batchErr := runtime.dependencies.Items.Batch(scoped, scope, ids)
			snapshots = make([]httpItemView, len(items))
			for index, item := range items {
				snapshots[index] = httpItem(item)
			}
			return batchErr
		})
	return documentshttp.Response{Status: http.StatusOK,
		Data: map[string]any{"items": snapshots}}, mapHTTPApplicationError(err)
}
