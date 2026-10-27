package usage

import (
	"context"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	"github.com/google/uuid"
)

type queryRows interface {
	Close()
	Err() error
	Next() bool
	Scan(...any) error
}

type queryMany func(context.Context, string, ...any) (queryRows, error)

type queryStore struct {
	tenantID uuid.UUID
	queryRow queryRow
	query    queryMany
	allowed  scopeCheck
	notFound rowNotFound
}

var _ usageapp.CurrentUsageStore = (*queryStore)(nil)

func (storage *queryStore) queryUsable(ctx context.Context) bool {
	return storage != nil && storage.tenantID != uuid.Nil && storage.queryRow != nil &&
		storage.query != nil && storage.allowed != nil && storage.notFound != nil &&
		ctx != nil && ctx.Err() == nil && storage.allowed(ctx)
}
