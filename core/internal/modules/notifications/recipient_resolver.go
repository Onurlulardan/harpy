package notifications

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type recipientResolver struct{ binder *postgres.Binder }

func (resolver recipientResolver) ResolveRecipient(ctx context.Context, scope api.Scope,
	recipientID ports.ID,
) (ports.Recipient, error) {
	return resolver.binder.Recipient(ctx, scope, recipientID)
}

var _ ports.RecipientResolver = recipientResolver{}
