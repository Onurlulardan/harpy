package folder

import (
	"context"
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Create(ctx context.Context, scope platformtx.TenantScope,
	request CreateRequest,
) (domain.Folder, error) {
	request.DisplayName = strings.TrimSpace(request.DisplayName)
	if !foundation.ValidIdempotency(request.IdempotencyKey) {
		return domain.Folder{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, Command, OperationCreate, PermissionCreate)
	if err != nil {
		return domain.Folder{}, err
	}
	folderID, err := application.ids.NewUUIDv7()
	if err != nil || !folderID.UUIDv7() {
		return domain.Folder{}, ErrUnavailable
	}
	now := application.clock.Now()
	if _, err = domain.NewFolder(folderID, identity.TenantID, request.ParentID,
		identity.ActorID, request.DisplayName, now); err != nil {
		return domain.Folder{}, ErrInvalid
	}
	return application.createBound(ctx, store, identity, request, folderID, now)
}

func (application *Application) createBound(ctx context.Context, store Store,
	identity foundation.Identity, request CreateRequest, folderID domain.ID, now time.Time,
) (domain.Folder, error) {
	mutation, err := application.newMutation(identity.ActorID, folderID, request.ParentID,
		0, request.DisplayName, request.IdempotencyKey, OperationCreate, now)
	if err != nil {
		return domain.Folder{}, err
	}
	result, replay, err := store.Create(ctx, mutation)
	if err != nil || !replay && result.ID != folderID {
		return domain.Folder{}, mapStoreError(err)
	}
	return result, nil
}
