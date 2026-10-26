package formrecordarchive

import (
	"context"

	formrecordarchiveapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordarchive"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Get(ctx context.Context,
	access formrecordarchiveapi.Access, request formrecordarchiveapi.GetRequest,
) (formrecordarchiveapi.Snapshot, error) {
	if application == nil || !request.Valid() {
		return formrecordarchiveapi.Snapshot{}, formrecordarchiveapi.ErrInvalid
	}
	items := make([]domain.ID, len(request.ItemIDs))
	for index, itemID := range request.ItemIDs {
		items[index] = domain.ID(itemID)
	}
	digest, valid := archiveapp.RequestDigest(archiveapp.CreateRequest{
		ItemIDs: items, RetryOfArchiveID: domain.ID(request.RetryOfArchiveID),
		IdempotencyKey: "validated-by-caller"})
	if !valid || digest != request.RequestSHA256 {
		return formrecordarchiveapi.Snapshot{}, formrecordarchiveapi.ErrInvalid
	}
	granted, err := application.authorize(ctx, access, request.ItemIDs,
		archiveapp.OperationGet)
	if err != nil {
		return formrecordarchiveapi.Snapshot{}, err
	}
	scopeRequest, err := application.snapshotScope(granted.context, granted.actorID,
		granted.tenantID, archiveapp.OperationGet)
	if err != nil {
		return formrecordarchiveapi.Snapshot{}, err
	}
	var result archiveapp.Snapshot
	err = application.dependencies.Scopes.InTenant(granted.context, scopeRequest,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var getErr error
			result, getErr = application.dependencies.Archives.Get(scoped, scope,
				archiveapp.GetRequest{ArchiveID: domain.ID(request.ArchiveID)})
			return getErr
		})
	if err != nil {
		return formrecordarchiveapi.Snapshot{}, mapError(err)
	}
	if result.RequestSHA256 != request.RequestSHA256 {
		return formrecordarchiveapi.Snapshot{}, formrecordarchiveapi.ErrNotFound
	}
	return project(result, granted.tenantID, granted.actorID)
}
