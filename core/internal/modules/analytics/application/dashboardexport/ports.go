package dashboardexport

import (
	"context"
	"io"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type ScopeAccess uint8

const (
	ScopeCommand ScopeAccess = iota + 1
	ScopeSnapshot
)

type ScopeIdentity struct{ ActorID, TenantID uuid.UUID }

type Store interface {
	Submit(context.Context, SubmitMutation) (Snapshot, bool, error)
	Get(context.Context, uuid.UUID, uuid.UUID, time.Time) (Snapshot, error)
	RequestDelete(context.Context, DeleteMutation) (Snapshot, bool, bool, error)
	DescribeContent(context.Context, uuid.UUID, uuid.UUID, time.Time) (StoredContent, error)
}

type StoreBinder interface {
	Bind(context.Context, platformtx.TenantScope, ScopeAccess) (Store, ScopeIdentity, error)
}

type Authorizer interface {
	Authorize(context.Context, platformtx.TenantScope, Permission) error
}

type IDSource interface{ NewUUIDv7() (uuid.UUID, error) }
type Clock interface{ Now() time.Time }

type Scheduler interface {
	Schedule(context.Context, platformtx.TenantScope, Job) error
}

type ContentRuntime interface {
	Open(context.Context, StoredContent) (io.ReadCloser, error)
}

type ReadCloser = io.ReadCloser
