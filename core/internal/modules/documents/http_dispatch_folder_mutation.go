package documents

import (
	"context"
	"net/http"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	folderapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/folder"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) mutateHTTPFolder(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	id, err := parseHTTPID(request.ResourceID, false)
	if err != nil || !onlyHTTPQuery(request.Query) {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	var result folderapp.Folder
	err = runtime.command(ctx, request.Access, request.Operation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			switch request.Operation {
			case folderapp.OperationUpdate:
				input, decodeErr := decodeHTTPRequest[httpUpdateFolderRequest](request.Document)
				if decodeErr != nil {
					return decodeErr
				}
				result, operationErr = runtime.dependencies.Folders.Update(scoped, scope,
					folderapp.UpdateRequest{FolderID: id, ExpectedVersion: input.ExpectedVersion,
						DisplayName: input.Name, IdempotencyKey: request.IdempotencyKey})
			case folderapp.OperationMove:
				input, decodeErr := decodeHTTPRequest[httpMoveFolderRequest](request.Document)
				if decodeErr != nil {
					return decodeErr
				}
				parent := ""
				if input.ParentID != nil {
					parent = *input.ParentID
				}
				parentID, parseErr := parseHTTPID(parent, true)
				if parseErr != nil {
					return parseErr
				}
				result, operationErr = runtime.dependencies.Folders.Move(scoped, scope,
					folderapp.MoveRequest{FolderID: id, ParentID: parentID,
						ExpectedVersion: input.ExpectedVersion,
						IdempotencyKey:  request.IdempotencyKey})
			case folderapp.OperationDelete, folderapp.OperationRestore:
				input, decodeErr := decodeHTTPRequest[httpRevisionRequest](request.Document)
				if decodeErr != nil {
					return decodeErr
				}
				mutation := folderapp.DeleteRequest{FolderID: id,
					ExpectedVersion: input.ExpectedVersion,
					IdempotencyKey:  request.IdempotencyKey}
				if request.Operation == folderapp.OperationDelete {
					result, operationErr = runtime.dependencies.Folders.Delete(scoped, scope, mutation)
				} else {
					result, operationErr = runtime.dependencies.Folders.Restore(scoped, scope, mutation)
				}
			}
			return operationErr
		})
	return documentshttp.Response{Status: http.StatusOK, Data: httpFolder(result)},
		mapHTTPApplicationError(err)
}
