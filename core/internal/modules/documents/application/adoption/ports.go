package adoption

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Store interface {
	CreateRun(context.Context, Authority, string, time.Time) (Run, bool, error)
	Inventory(context.Context, InventoryRequest) (Checkpoint, bool, error)
	PrepareCopy(context.Context, Authority, domain.ID, time.Time) (Checkpoint, error)
	VerifyCopy(context.Context, Authority, Checkpoint, CopyReceipt) (Checkpoint, error)
	BeginShadow(context.Context, Authority, time.Time) (Run, error)
	Reconcile(context.Context, ReconcileRequest) (Run, error)
	Accept(context.Context, Authority, [32]byte, string, time.Time) (Run, error)
	RecordHit(context.Context, CompatibilityHit) (CompatibilityHit, error)
	RequestReverse(context.Context, ReverseRequest) error
	DecideReverse(context.Context, ReverseDecision) error
}

type Copier interface {
	Copy(context.Context, Authority, Checkpoint) (CopyReceipt, error)
}

type JobStoreBinder interface {
	BindJob(context.Context, platformtx.JobTenantScope) (Store, error)
}

type ScopeStoreBinder interface {
	Bind(context.Context, platformtx.TenantScope) (Store, error)
}

type JobRunner interface {
	InRepeatableReadJobTenant(context.Context, domain.ID,
		func(context.Context, platformtx.JobTenantScope) error) error
}

type Clock interface{ Now() time.Time }
