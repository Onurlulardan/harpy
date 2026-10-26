package offering

import (
	"context"

	offeringapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/offering"
)

type row interface{ Scan(...any) error }
type queryRow func(context.Context, string, ...any) row
type execute func(context.Context, string, ...any) error
type scopeCheck func(context.Context) bool

type store struct {
	queryRow queryRow
	execute  execute
	allowed  scopeCheck
}

var _ offeringapp.Store = (*store)(nil)

func (storage *store) usable(ctx context.Context) bool {
	return storage != nil && storage.queryRow != nil && storage.execute != nil &&
		storage.allowed != nil && ctx != nil && ctx.Err() == nil && storage.allowed(ctx)
}
