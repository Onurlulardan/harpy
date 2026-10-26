package formscleanup

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Store interface {
	PersistRequest(context.Context, Mutation) (Snapshot, error)
	CaptureLinks(context.Context, domain.ID, time.Time) (Snapshot, []Candidate, error)
	StageEnqueue(context.Context, EnqueuePlan) (Candidate, error)
	AcknowledgeEnqueue(context.Context, Enqueued) error
	SealReceipt(context.Context, domain.ID, time.Time) (Snapshot, error)
}

type StoreBinder interface {
	Bind(context.Context, platformtx.TenantScope) (Store, error)
}

type Jobs interface {
	EnqueuePurge(context.Context, platformtx.TenantScope,
		foundation.ID, foundation.ID, time.Time, []byte) error
}
