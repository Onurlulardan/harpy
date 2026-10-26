package purge

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Store interface {
	Lock(context.Context, domain.ID, time.Time) (Candidate, error)
	Begin(context.Context, Mutation) (Result, error)
	PersistCompletion(context.Context, Mutation) (Result, error)
	ScanPending(context.Context, domain.ID, domain.ID, uint16) (DeletionPage, error)
}

type StoreBinder interface {
	BindJob(context.Context, platformtx.JobTenantScope) (Store, error)
}

type Jobs interface {
	EnqueueDeletion(context.Context, platformtx.JobTenantScope,
		domain.ID, domain.ID, []byte) error
}

type JobRunner interface {
	InRepeatableReadJobTenant(context.Context, domain.ID,
		func(context.Context, platformtx.JobTenantScope) error) error
}
