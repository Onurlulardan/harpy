package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (application *Application) MarkRead(ctx context.Context, scope api.Scope,
	id domain.ID,
) (domain.Message, error) {
	if !id.Valid() {
		return domain.Message{}, ErrInvalid
	}
	stores, err := application.bindOperation(ctx, scope, PermissionInboxUpdate)
	if err != nil {
		return domain.Message{}, err
	}
	_, recipient, _ := scopeIDs(scope)
	message, err := stores.Messages().GetMessage(ctx, id, recipient)
	if err != nil {
		return domain.Message{}, ErrNotFound
	}
	updated, err := message.MarkRead(message.Revision, application.clock.Now())
	if err != nil {
		return domain.Message{}, ErrConflict
	}
	if updated.Revision != message.Revision {
		if err = stores.Messages().PersistMessageRevision(ctx, updated, message.Revision); err != nil {
			return domain.Message{}, ErrConflict
		}
		if err = stores.Evidence().AppendAudit(ctx, "notifications.inbox.read",
			id, recipient, updated.Revision); err != nil {
			return domain.Message{}, ErrUnavailable
		}
	}
	return updated, nil
}
