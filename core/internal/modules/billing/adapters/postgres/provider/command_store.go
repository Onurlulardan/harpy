package provider

import (
	"context"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

type CommandEffects interface {
	EnqueueCommandRoute(context.Context, providerapp.CommandSnapshot) error
}

type commandStore struct {
	tenantID uuid.UUID
	actorID  uuid.UUID
	database databaseAccess
	effects  CommandEffects
}

func (storage *commandStore) usable(ctx context.Context) bool {
	return storage != nil && storage.tenantID != uuid.Nil &&
		storage.actorID != uuid.Nil && storage.database.queryRow != nil &&
		storage.database.allowed != nil && storage.effects != nil &&
		ctx != nil && ctx.Err() == nil && storage.database.allowed(ctx)
}

var _ providerapp.CommandStore = (*commandStore)(nil)
