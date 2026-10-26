package documents

import (
	"context"
	"net/http"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) mutateHTTPItem(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	itemID, err := parseHTTPID(request.ResourceID, false)
	if err != nil || !onlyHTTPQuery(request.Query) {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	var result itemapp.Snapshot
	err = runtime.command(ctx, request.Access, request.Operation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var operationErr error
			switch request.Operation {
			case itemapp.OperationUpdate:
				input, decodeErr := decodeHTTPRequest[httpUpdateItemRequest](request.Document)
				if decodeErr != nil {
					return decodeErr
				}
				result, operationErr = runtime.dependencies.Items.Update(scoped, scope,
					itemapp.MetadataRequest{ItemID: itemID, RowVersion: input.ExpectedVersion,
						Description: input.Description, Tags: input.Tags,
						IdempotencyKey: request.IdempotencyKey})
			case itemapp.OperationMove:
				input, decodeErr := decodeHTTPRequest[httpMoveItemRequest](request.Document)
				if decodeErr != nil {
					return decodeErr
				}
				folderID, parseErr := parseHTTPID(input.FolderID, false)
				if parseErr != nil {
					return parseErr
				}
				result, operationErr = runtime.dependencies.Items.Move(scoped, scope,
					itemapp.MoveRequest{ItemID: itemID, FolderID: folderID,
						RowVersion:     input.ExpectedVersion,
						IdempotencyKey: request.IdempotencyKey})
			case itemapp.OperationPurge:
				input, decodeErr := decodeHTTPRequest[httpPurgeItemRequest](request.Document)
				if decodeErr != nil {
					return decodeErr
				}
				result, operationErr = runtime.dependencies.Items.RequestPurge(scoped, scope,
					itemapp.PurgeRequest{MutationRequest: itemapp.MutationRequest{ItemID: itemID,
						RowVersion:     input.ExpectedVersion,
						IdempotencyKey: request.IdempotencyKey}, Reason: input.Reason})
			default:
				input, decodeErr := decodeHTTPRequest[httpRevisionRequest](request.Document)
				if decodeErr != nil {
					return decodeErr
				}
				mutation := itemapp.MutationRequest{ItemID: itemID,
					RowVersion: input.ExpectedVersion, IdempotencyKey: request.IdempotencyKey}
				switch request.Operation {
				case itemapp.OperationDelete:
					result, operationErr = runtime.dependencies.Items.Delete(scoped, scope, mutation)
				case itemapp.OperationRestore:
					result, operationErr = runtime.dependencies.Items.Restore(scoped, scope, mutation)
				case itemapp.OperationCancelPurge:
					result, operationErr = runtime.dependencies.Items.CancelPurge(scoped, scope, mutation)
				}
			}
			return operationErr
		})
	status := http.StatusOK
	if request.Operation == itemapp.OperationPurge {
		status = http.StatusAccepted
	}
	return documentshttp.Response{Status: status, Data: httpItem(result)},
		mapHTTPApplicationError(err)
}
