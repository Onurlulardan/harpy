package uploadreconcile

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Store interface {
	Prepare(context.Context, domain.ID, uint64, time.Time) (Plan, error)
	PersistFinalization(context.Context, Plan, string, domain.ID, time.Time) (Finalization, error)
}

type StoreBinder interface {
	BindJob(context.Context, platformtx.JobTenantScope) (Store, error)
}

type Inventory interface {
	FindExactVersion(context.Context, string, uint64, int64) (string, bool, error)
}

type Billing interface {
	Release(context.Context, platformtx.JobTenantScope, domain.ID, domain.ID) error
}

type Jobs interface {
	EnqueueDeletion(context.Context, platformtx.JobTenantScope,
		domain.ID, domain.ID, []byte) error
}

type JobRunner interface {
	InRepeatableReadJobTenant(context.Context, domain.ID,
		func(context.Context, platformtx.JobTenantScope) error) error
}
