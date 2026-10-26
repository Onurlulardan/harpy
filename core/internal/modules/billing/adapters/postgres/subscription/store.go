package subscription

import (
	"context"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

type row interface{ Scan(...any) error }
type queryRow func(context.Context, string, ...any) row
type execute func(context.Context, string, ...any) error
type scopeCheck func(context.Context) bool
type rowNotFound func(error) bool

type bindingMode string

const (
	modeActor    bindingMode = "actor"
	modeProvider bindingMode = "provider"
	modeJob      bindingMode = "job"
)

type store struct {
	tenantID uuid.UUID
	actorID  uuid.UUID
	mode     bindingMode
	queryRow queryRow
	execute  execute
	allowed  scopeCheck
	notFound rowNotFound
	effects  Effects
}

var _ subscriptionapp.Store = (*store)(nil)

func (storage *store) usable(ctx context.Context, mode bindingMode) bool {
	return storage != nil && storage.tenantID != uuid.Nil && storage.mode == mode &&
		storage.queryRow != nil && storage.execute != nil && storage.allowed != nil &&
		storage.notFound != nil && storage.effects != nil && ctx != nil &&
		ctx.Err() == nil && storage.allowed(ctx)
}
