package discounteligibility

import (
	"context"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	"github.com/google/uuid"
)

type row interface{ Scan(...any) error }
type rows interface {
	Close()
	Err() error
	Next() bool
	Scan(...any) error
}
type queryRow func(context.Context, string, ...any) row
type queryMany func(context.Context, string, ...any) (rows, error)
type scopeCheck func(context.Context) bool
type rowNotFound func(error) bool

type store struct {
	tenantID uuid.UUID
	queryRow queryRow
	query    queryMany
	allowed  scopeCheck
	notFound rowNotFound
}

var _ discountapp.EligibilityStore = (*store)(nil)

func (storage *store) usable(ctx context.Context) bool {
	return storage != nil && storage.tenantID != uuid.Nil && storage.queryRow != nil &&
		storage.query != nil && storage.allowed != nil && storage.notFound != nil &&
		ctx != nil && ctx.Err() == nil && storage.allowed(ctx)
}
