package application

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type ScopeIdentity struct{ ActorID, TenantID uuid.UUID }

type AppendMutation struct {
	Entry         domain.Entry
	RequestSHA256 [32]byte
}

type Store interface {
	Append(context.Context, AppendMutation) (domain.Entry, bool, error)
	List(context.Context, domain.Filter) (domain.Page, error)
	Stats(context.Context, time.Time) (domain.Stats, error)
	Get(context.Context, domain.ID) (domain.Entry, error)
}

type StoreBinder interface {
	Bind(context.Context, platformtx.TenantScope, ScopeAccess) (Store, ScopeIdentity, error)
}

type Authorizer interface {
	Authorize(context.Context, platformtx.TenantScope, Permission) error
}

type IDSource interface{ NewUUIDv7() (uuid.UUID, error) }
type Clock interface{ Now() time.Time }
