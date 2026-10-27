package folder

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Update(ctx context.Context, scope platformtx.TenantScope,
	request UpdateRequest,
) (domain.Folder, error) {
	request.DisplayName = strings.TrimSpace(request.DisplayName)
	if !request.FolderID.UUIDv7() || request.ExpectedVersion == 0 ||
		!foundation.ValidIdempotency(request.IdempotencyKey) {
		return domain.Folder{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, Command, OperationUpdate, PermissionUpdate)
	if err != nil {
		return domain.Folder{}, err
	}
	mutation, err := application.newMutation(identity.ActorID, request.FolderID, domain.ID{},
		request.ExpectedVersion, request.DisplayName, request.IdempotencyKey, OperationUpdate,
		application.clock.Now())
	if err != nil {
		return domain.Folder{}, err
	}
	return updateBound(ctx, store, mutation)
}

func updateBound(ctx context.Context, store Store, mutation Mutation) (domain.Folder, error) {
	return applyBoundMutation(ctx, store.Update, mutation)
}
