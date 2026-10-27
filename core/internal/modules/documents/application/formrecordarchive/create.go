package formrecordarchive

import (
	"context"

	formrecordarchiveapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordarchive"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Create(ctx context.Context,
	access formrecordarchiveapi.Access, request formrecordarchiveapi.CreateRequest,
) (formrecordarchiveapi.Snapshot, error) {
	if application == nil || !request.Valid() {
		return formrecordarchiveapi.Snapshot{}, formrecordarchiveapi.ErrInvalid
	}
	granted, err := application.authorize(ctx, access, request.ItemIDs,
		archiveapp.OperationCreate)
	if err != nil {
		return formrecordarchiveapi.Snapshot{}, err
	}
	scopeRequest, err := application.commandScope(granted.context, granted.actorID,
		granted.tenantID, archiveapp.OperationCreate)
	if err != nil {
		return formrecordarchiveapi.Snapshot{}, err
	}
	definitionID, recordID := domain.ID(request.DefinitionID), domain.ID(request.RecordID)
	retryID := domain.ID(request.RetryOfArchiveID)
	var result archiveapp.Snapshot
	err = application.dependencies.Scopes.InTenant(granted.context, scopeRequest,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			if matchErr := application.dependencies.Links.Match(scoped, scope,
				granted.tenantID, granted.actorID, definitionID, recordID,
				granted.itemIDs); matchErr != nil {
				return matchErr
			}
			var createErr error
			result, createErr = application.dependencies.Archives.Create(scoped, scope,
				archiveapp.CreateRequest{ItemIDs: granted.itemIDs,
					RetryOfArchiveID: retryID, IdempotencyKey: formRecordIdempotency(
						request.IdempotencyKey, definitionID, recordID)})
			return createErr
		})
	if err != nil {
		return formrecordarchiveapi.Snapshot{}, mapError(err)
	}
	return project(result, granted.tenantID, granted.actorID)
}
