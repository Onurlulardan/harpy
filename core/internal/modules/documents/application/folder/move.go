package folder

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Move(ctx context.Context, scope platformtx.TenantScope,
	request MoveRequest,
) (domain.Folder, error) {
	if !request.FolderID.UUIDv7() || request.ParentID.Valid() && !request.ParentID.UUIDv7() ||
		request.ParentID == request.FolderID || request.ExpectedVersion == 0 ||
		!foundation.ValidIdempotency(request.IdempotencyKey) {
		return domain.Folder{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, Command, OperationMove, PermissionMove)
	if err != nil {
		return domain.Folder{}, err
	}
	mutation, err := application.newMutation(identity.ActorID, request.FolderID, request.ParentID,
		request.ExpectedVersion, "", request.IdempotencyKey, OperationMove, application.clock.Now())
	if err != nil {
		return domain.Folder{}, err
	}
	return moveBound(ctx, store, mutation)
}

func moveBound(ctx context.Context, store Store, mutation Mutation) (domain.Folder, error) {
	return applyBoundMutation(ctx, store.Move, mutation)
}
