package resourceentitlement

import (
	"context"

	resourceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/resourceentitlement"
	"github.com/google/uuid"
)

type row interface{ Scan(...any) error }
type queryRow func(context.Context, string, ...any) row
type execute func(context.Context, string, ...any) error
type scopeCheck func(context.Context) bool

type store struct {
	tenantID uuid.UUID
	actorID  uuid.UUID
	queryRow queryRow
	execute  execute
	allowed  scopeCheck
}

var _ resourceapp.Store = (*store)(nil)
var _ resourceapp.ReconciliationStore = (*store)(nil)

func (storage *store) usable(ctx context.Context) bool {
	return storage != nil && storage.tenantID != uuid.Nil && storage.actorID != uuid.Nil &&
		storage.queryRow != nil && storage.execute != nil && storage.allowed != nil &&
		ctx != nil && ctx.Err() == nil && storage.allowed(ctx)
}
