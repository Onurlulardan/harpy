package adoption

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type ScopeAccess uint8

const (
	ScopeCommand ScopeAccess = iota + 1
	ScopeSnapshot
)

type Identity struct{ ActorID, TenantID uuid.UUID }

type Binder interface {
	Bind(context.Context, platformtx.TenantScope, ScopeAccess) (Store, Identity, error)
}

type Store interface {
	ValidateInventory(context.Context) (Inventory, error)
	FinalizeRun(context.Context, Run) (Run, error)
	RecordHit(context.Context, RouteRequest, time.Time) error
	AcceptConsumer(context.Context, ConsumerAck, time.Time) error
	Retire(context.Context, string, time.Time) error
}

type Authorizer interface {
	Authorize(context.Context, platformtx.TenantScope, string) error
}

type Clock interface{ Now() time.Time }
