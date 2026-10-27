package documents

import (
	"context"
	"net/http"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) getHTTPItem(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	id, err := parseHTTPID(request.ResourceID, false)
	if err != nil || !onlyHTTPQuery(request.Query) {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	var result itemapp.Snapshot
	err = runtime.snapshot(ctx, request.Access, request.Operation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var getErr error
			result, getErr = runtime.dependencies.Items.Get(scoped, scope,
				itemapp.GetRequest{ItemID: id})
			return getErr
		})
	return documentshttp.Response{Status: http.StatusOK, Data: httpItem(result)},
		mapHTTPApplicationError(err)
}

func (runtime *HTTPRuntime) listHTTPItems(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	if !onlyHTTPQuery(request.Query, "folder_id", "after_id", "after_updated_at", "limit") {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	folder, folderOK := httpQueryValue(request.Query, "folder_id")
	after, afterOK := httpQueryValue(request.Query, "after_id")
	folderID, folderErr := parseHTTPID(folder, false)
	afterID, afterErr := parseHTTPID(after, true)
	afterAt, timeErr := httpTime(request.Query, "after_updated_at")
	limit, limitErr := httpLimit(request.Query, 100)
	if !folderOK || !afterOK || folderErr != nil || afterErr != nil || timeErr != nil ||
		limitErr != nil || afterID.Valid() != !afterAt.IsZero() {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	var result itemapp.Page
	err := runtime.snapshot(ctx, request.Access, request.Operation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var listErr error
			result, listErr = runtime.dependencies.Items.List(scoped, scope,
				itemapp.ListRequest{FolderID: folderID,
					After: itemapp.Cursor{ID: afterID, UpdatedAt: afterAt}, Limit: limit})
			return listErr
		})
	views := make([]httpItemView, len(result.Items))
	for index, item := range result.Items {
		views[index] = httpItem(item)
	}
	data := map[string]any{"items": views, "next_id": result.Next.ID.String(),
		"next_updated_at": result.Next.UpdatedAt}
	return documentshttp.Response{Status: http.StatusOK, Data: data},
		mapHTTPApplicationError(err)
}
