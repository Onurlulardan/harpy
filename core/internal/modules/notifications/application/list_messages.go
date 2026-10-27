package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (application *Application) ListMessages(ctx context.Context, scope api.Scope,
	filter ports.MessageFilter,
) (ports.MessagePage, error) {
	if filter.Limit < 1 || filter.Limit > 100 ||
		filter.Category != nil && !filter.Category.Valid() {
		return ports.MessagePage{}, ErrInvalid
	}
	stores, err := application.bindOperation(ctx, scope, PermissionInboxRead)
	if err != nil {
		return ports.MessagePage{}, err
	}
	_, recipient, err := scopeIDs(scope)
	if err != nil {
		return ports.MessagePage{}, ErrInvalid
	}
	page, err := stores.Messages().ListMessages(ctx, recipient, filter)
	if err != nil {
		return ports.MessagePage{}, ErrUnavailable
	}
	for _, message := range page.Items {
		if message.RecipientID != recipient {
			return ports.MessagePage{}, ErrUnavailable
		}
	}
	return page, nil
}
