package deletion

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Store interface {
	Lease(context.Context, domain.ID, domain.ID, time.Time, time.Duration) (Task, error)
	BeginReconcile(context.Context, Task) error
	Confirm(context.Context, Task, DeleteReceipt, domain.ID, time.Time) (Result, error)
}

type StoreBinder interface {
	BindJob(context.Context, platformtx.JobTenantScope) (Store, error)
}

type ObjectStore interface {
	DeleteExact(context.Context, Task) (DeleteReceipt, error)
}

type Billing interface {
	Decrease(context.Context, platformtx.JobTenantScope, domain.ID, int64, domain.ID) error
	Release(context.Context, platformtx.JobTenantScope, domain.ID, domain.ID) error
	Keep(context.Context, platformtx.JobTenantScope, domain.ID, domain.ID) error
}

type JobRunner interface {
	InRepeatableReadJobTenant(context.Context, domain.ID,
		func(context.Context, platformtx.JobTenantScope) error) error
}
