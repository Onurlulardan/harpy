package folder

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Delete(ctx context.Context, scope platformtx.TenantScope,
	request DeleteRequest,
) (domain.Folder, error) {
	if !request.FolderID.UUIDv7() || request.ExpectedVersion == 0 ||
		!foundation.ValidIdempotency(request.IdempotencyKey) {
		return domain.Folder{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, Command, OperationDelete, PermissionDelete)
	if err != nil {
		return domain.Folder{}, err
	}
	mutation, err := application.newMutation(identity.ActorID, request.FolderID, domain.ID{},
		request.ExpectedVersion, "", request.IdempotencyKey, OperationDelete, application.clock.Now())
	if err != nil {
		return domain.Folder{}, err
	}
	return deleteBound(ctx, store, mutation)
}

func deleteBound(ctx context.Context, store Store, mutation Mutation) (domain.Folder, error) {
	return applyBoundMutation(ctx, store.Delete, mutation)
}
