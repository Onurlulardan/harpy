package ports

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

// Scope is the consumer-local view of one still-active platform TenantScope.
// Adapters must reject values that are not backed by that exact scope.
type Scope interface {
	TenantID() domain.ID
	ActorID() domain.ID
	OperationID() string
	SnapshotQuery() bool
}

type UnitOfWork interface {
	Bind(context.Context, Scope) (Stores, error)
}

type Stores interface {
	Boards() BoardStore
	Tasks() TaskStore
	Collaboration() CollaborationStore
	Calendar() CalendarStore
	Projections() ProjectionStore
	Operations() OperationStore
}
