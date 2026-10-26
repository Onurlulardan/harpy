package formscleanup

import (
	"context"

	formscleanupapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formscleanup"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Request(ctx context.Context,
	scope platformtx.TenantScope, request formscleanupapi.Request,
) (formscleanupapi.Progress, error) {
	operationID, operationErr := foundation.ParseID(request.OperationID)
	definitionID, definitionErr := foundation.ParseID(request.DefinitionID)
	requestedAt := request.RequestedAt.UTC().Truncate(1e3)
	if application == nil || operationErr != nil || definitionErr != nil ||
		!operationID.UUIDv7() || !definitionID.UUIDv7() || request.Epoch < 2 ||
		request.RequestedAt != requestedAt || !foundation.ValidInstant(requestedAt) {
		return formscleanupapi.Progress{}, ErrInvalid
	}
	store, err := application.bind(ctx, scope)
	if err != nil {
		return formscleanupapi.Progress{}, err
	}
	snapshot, err := store.PersistRequest(ctx, Mutation{OperationID: operationID,
		DefinitionID: definitionID, Epoch: request.Epoch, RequestedAt: requestedAt})
	if err != nil {
		return formscleanupapi.Progress{}, mapError(err)
	}
	if !snapshotMatches(snapshot, operationID, definitionID, request.Epoch, requestedAt) {
		return formscleanupapi.Progress{}, ErrConflict
	}
	return application.advance(ctx, scope, store, operationID)
}
