package invoice

import (
	"context"

	invoiceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/invoice"
	"github.com/google/uuid"
)

type row interface{ Scan(...any) error }
type queryRow func(context.Context, string, ...any) row
type execute func(context.Context, string, ...any) error
type scopeCheck func(context.Context) bool
type rowNotFound func(error) bool

type store struct {
	tenantID uuid.UUID
	queryRow queryRow
	execute  execute
	allowed  scopeCheck
	notFound rowNotFound
	effects  Effects
}

var _ invoiceapp.Store = (*store)(nil)

func (storage *store) usable(ctx context.Context) bool {
	return storage != nil && storage.tenantID != uuid.Nil && storage.queryRow != nil &&
		storage.execute != nil && storage.allowed != nil && storage.notFound != nil &&
		storage.effects != nil &&
		ctx != nil && ctx.Err() == nil && storage.allowed(ctx)
}
