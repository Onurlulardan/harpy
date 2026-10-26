package documents

import (
	"context"
	"net/http"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type httpCreateFolderRequest struct {
	ParentID *string `json:"parent_id"`
	Name     string  `json:"name"`
}

type httpUpdateFolderRequest struct {
	ExpectedVersion uint64 `json:"expected_version"`
	Name            string `json:"name"`
}

type httpMoveFolderRequest struct {
	ExpectedVersion uint64  `json:"expected_version"`
	ParentID        *string `json:"parent_id"`
}

type httpRevisionRequest struct {
	ExpectedVersion uint64 `json:"expected_version"`
}

func (runtime *HTTPRuntime) dispatchFolder(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	switch request.Operation {
	case folderapp.OperationCreate:
		return runtime.createHTTPFolder(ctx, request)
	case folderapp.OperationGet:
		return runtime.getHTTPFolder(ctx, request)
	case folderapp.OperationList:
		return runtime.listHTTPFolders(ctx, request)
	case folderapp.OperationUpdate, folderapp.OperationMove, folderapp.OperationDelete,
		folderapp.OperationRestore:
		return runtime.mutateHTTPFolder(ctx, request)
	case folderapp.OperationContents:
		return runtime.listHTTPFolderContents(ctx, request)
	default:
		return documentshttp.Response{}, documentshttp.ErrForbidden
	}
}

func (runtime *HTTPRuntime) createHTTPFolder(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	input, err := decodeHTTPRequest[httpCreateFolderRequest](request.Document)
	if err != nil || !onlyHTTPQuery(request.Query) {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	parent := ""
	if input.ParentID != nil {
		parent = *input.ParentID
	}
	parentID, err := parseHTTPID(parent, true)
	if err != nil {
		return documentshttp.Response{}, err
	}
	var result any
	err = runtime.command(ctx, request.Access, request.Operation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			folder, createErr := runtime.dependencies.Folders.Create(scoped, scope,
				folderapp.CreateRequest{ParentID: parentID, DisplayName: input.Name,
					IdempotencyKey: request.IdempotencyKey})
			result = httpFolder(folder)
			return createErr
		})
	if err != nil {
		return documentshttp.Response{}, mapHTTPApplicationError(err)
	}
	view := result.(httpFolderView)
	return documentshttp.Response{Status: http.StatusCreated, Data: view,
		Location: "/api/folders/" + view.ID}, nil
}

func (runtime *HTTPRuntime) getHTTPFolder(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	id, err := parseHTTPID(request.ResourceID, false)
	if err != nil || !onlyHTTPQuery(request.Query) {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	var result httpFolderView
	err = runtime.snapshot(ctx, request.Access, request.Operation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			folder, getErr := runtime.dependencies.Folders.Get(scoped, scope,
				folderapp.GetRequest{FolderID: id})
			result = httpFolder(folder)
			return getErr
		})
	return documentshttp.Response{Status: http.StatusOK, Data: result},
		mapHTTPApplicationError(err)
}

func (runtime *HTTPRuntime) listHTTPFolders(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	if !onlyHTTPQuery(request.Query, "parent_id", "after_id", "limit") {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	parent, valid := httpQueryValue(request.Query, "parent_id")
	after, afterValid := httpQueryValue(request.Query, "after_id")
	parentID, parentErr := parseHTTPID(parent, true)
	afterID, afterErr := parseHTTPID(after, true)
	limit, limitErr := httpLimit(request.Query, 200)
	if !valid || !afterValid || parentErr != nil || afterErr != nil || limitErr != nil {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	var result folderapp.Page
	err := runtime.snapshot(ctx, request.Access, request.Operation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var listErr error
			result, listErr = runtime.dependencies.Folders.List(scoped, scope,
				folderapp.ListRequest{ParentID: parentID, AfterID: afterID, Limit: limit})
			return listErr
		})
	views := make([]httpFolderView, len(result.Folders))
	for index, folder := range result.Folders {
		views[index] = httpFolder(folder)
	}
	data := map[string]any{"items": views, "next_id": result.NextID.String()}
	return documentshttp.Response{Status: http.StatusOK, Data: data},
		mapHTTPApplicationError(err)
}
