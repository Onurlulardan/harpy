package documents

import (
	"context"
	"net/http"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) listHTTPFolderContents(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	folderID, err := parseHTTPID(request.ResourceID, false)
	if err != nil || !onlyHTTPQuery(request.Query, "limit") {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	limit, err := httpLimit(request.Query, 100)
	if err != nil {
		return documentshttp.Response{}, err
	}
	var folders folderapp.Page
	var items itemapp.Page
	err = runtime.snapshot(ctx, request.Access, request.Operation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var listErr error
			folders, listErr = runtime.dependencies.Folders.Contents(scoped, scope,
				folderapp.ListRequest{ParentID: folderID, Limit: limit})
			if listErr != nil {
				return listErr
			}
			items, listErr = runtime.dependencies.Items.FolderContents(scoped, scope,
				itemapp.ListRequest{FolderID: folderID, Limit: limit})
			return listErr
		})
	folderViews := make([]httpFolderView, len(folders.Folders))
	for index, folder := range folders.Folders {
		folderViews[index] = httpFolder(folder)
	}
	itemViews := make([]httpItemView, len(items.Items))
	for index, item := range items.Items {
		itemViews[index] = httpItem(item)
	}
	data := map[string]any{"folders": folderViews, "documents": itemViews}
	return documentshttp.Response{Status: http.StatusOK, Data: data},
		mapHTTPApplicationError(err)
}
