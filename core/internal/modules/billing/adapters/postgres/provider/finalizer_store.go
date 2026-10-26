package provider

import (
	"context"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

type finalizerStore struct {
	tenantID uuid.UUID
	database databaseAccess
}

func (storage *finalizerStore) usable(ctx context.Context) bool {
	return storage != nil && storage.tenantID != uuid.Nil &&
		storage.database.queryRow != nil && storage.database.allowed != nil &&
		ctx != nil && ctx.Err() == nil && storage.database.allowed(ctx)
}

var _ providerapp.FinalizerStore = (*finalizerStore)(nil)
