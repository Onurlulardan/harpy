package subscription

import (
	"context"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

type queryStore struct {
	tenantID uuid.UUID
	queryRow queryRow
	allowed  scopeCheck
	notFound rowNotFound
}

var _ subscriptionapp.CurrentStore = (*queryStore)(nil)

func (storage *queryStore) queryUsable(ctx context.Context) bool {
	return storage != nil && storage.tenantID != uuid.Nil && storage.queryRow != nil &&
		storage.allowed != nil && storage.notFound != nil && ctx != nil &&
		ctx.Err() == nil && storage.allowed(ctx)
}
