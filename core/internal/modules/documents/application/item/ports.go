package item

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Access uint8

const (
	Command Access = iota + 1
	SnapshotAccess
)

type Store interface {
	Get(context.Context, domain.ID) (Snapshot, error)
	List(context.Context, domain.ID, Cursor, uint16) (Page, error)
	ReadVersions(context.Context, domain.ID, uint64, uint16) (VersionPage, error)
	Update(context.Context, Mutation) (Snapshot, bool, error)
	Move(context.Context, Mutation) (Snapshot, bool, error)
	Trash(context.Context, Mutation) (Snapshot, bool, error)
	Restore(context.Context, Mutation) (Snapshot, bool, error)
	RequestPurge(context.Context, Mutation) (Snapshot, bool, error)
	CancelPurge(context.Context, Mutation) (Snapshot, bool, error)
}

type StoreBinder interface {
	Bind(context.Context, platformtx.TenantScope, Access) (Store, foundation.Identity, error)
}

type StepUp interface {
	Verify(context.Context, platformtx.TenantScope, string, domain.ID) (time.Time, error)
}
