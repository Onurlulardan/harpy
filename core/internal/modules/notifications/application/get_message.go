package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (application *Application) GetMessage(ctx context.Context, scope api.Scope,
	id domain.ID,
) (domain.Message, error) {
	if !id.Valid() {
		return domain.Message{}, ErrInvalid
	}
	stores, err := application.bindOperation(ctx, scope, PermissionInboxRead)
	if err != nil {
		return domain.Message{}, err
	}
	_, recipient, err := scopeIDs(scope)
	if err != nil {
		return domain.Message{}, ErrInvalid
	}
	message, err := stores.Messages().GetMessage(ctx, id, recipient)
	if err != nil {
		return domain.Message{}, ErrNotFound
	}
	return message, nil
}
