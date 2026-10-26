package archive

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
	FindCreate(context.Context, string, [32]byte) (Snapshot, bool, error)
	Create(context.Context, CreateMutation) (Snapshot, error)
	Get(context.Context, domain.ID) (Snapshot, error)
	LoadPlan(context.Context, domain.ID) (Plan, error)
	Pin(context.Context, domain.ID, time.Time) (Plan, error)
	PersistEffect(context.Context, domain.ID, ArtifactReceipt) (Plan, error)
	PersistReady(context.Context, domain.ID, ArtifactReceipt) (Snapshot, error)
	Expire(context.Context, domain.ID, domain.ID, time.Time) (Snapshot, bool, error)
}

type StoreBinder interface {
	Bind(context.Context, platformtx.TenantScope, Access) (Store, foundation.Identity, error)
	BindJob(context.Context, platformtx.JobTenantScope) (Store, error)
}

type CommandAuthorizer interface {
	Authorize(context.Context, platformtx.TenantScope, string) (uint64, error)
}

type RefreshAuthorizer interface {
	Refresh(context.Context, domain.ID, domain.ID, uint64, []domain.ID) error
}

type CommandJobs interface {
	Enqueue(context.Context, platformtx.TenantScope, domain.ID, domain.ID, []byte) error
}

type JobControl interface {
	Check(context.Context, domain.ID) error
}

type Billing interface {
	Reserve(context.Context, platformtx.JobTenantScope,
		domain.ID, domain.ID, int64, time.Time) error
	StartEffect(context.Context, platformtx.JobTenantScope, domain.ID, domain.ID) error
	Commit(context.Context, platformtx.JobTenantScope, domain.ID, domain.ID, int64) error
}

type ArtifactWriter interface {
	UpperBound(Plan) (int64, error)
	Write(context.Context, Plan, func(int) error) (ArtifactReceipt, error)
	Recover(context.Context, Plan) (ArtifactReceipt, error)
}

type JobRunner interface {
	InRepeatableReadJobTenant(context.Context, domain.ID,
		func(context.Context, platformtx.JobTenantScope) error) error
}
