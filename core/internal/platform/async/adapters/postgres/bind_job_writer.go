package postgres

import (
	"context"
	"fmt"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type JobWriter interface {
	EnqueueJob(context.Context, jobsv1.Enqueue) (EnqueueResult, error)
}

func BindJobWriter(ctx context.Context, scope platformtx.JobTenantScope,
	authority jobsv1.Authority, catalogs platformasync.WriteCatalogs,
) (JobWriter, error) {
	stamp, err := jobAuthorityStamp(scope, authority)
	if err != nil || ctx == nil || ctx.Err() != nil ||
		!platformtx.ContextOwnsJobScope(ctx, scope) || platformtx.NetworkAllowed(ctx) {
		return nil, ErrAuthorityMismatch
	}
	sealedJobs, sealedProducers, err := catalogs.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("bind job async catalogs: %w", err)
	}
	jobs, err := jobsv1.NewCatalog(sealedJobs.Policies())
	if err != nil {
		return nil, fmt.Errorf("bind job catalog: %w", err)
	}
	producers, err := eventsv1.NewProducerCatalog(sealedProducers.Policies())
	if err != nil {
		return nil, fmt.Errorf("bind job producer catalog: %w", err)
	}
	binder, err := platformtx.NewJobOwnerBinder(storageOwner,
		func(database platformtx.JobBoundExecutor) JobWriter {
			return &repository{database: database, authority: stamp,
				jobs: jobs, producers: producers}
		})
	if err != nil {
		return nil, fmt.Errorf("create async job writer binder: %w", err)
	}
	writer, err := platformtx.BindJob(scope, binder)
	if err != nil || writer == nil {
		return nil, fmt.Errorf("bind async job writer: %w", err)
	}
	return writer, nil
}

func jobAuthorityStamp(scope platformtx.JobTenantScope,
	authority jobsv1.Authority,
) (authorityStamp, error) {
	scopeID, err := uuid.Parse(authority.ScopeID)
	if err != nil || authority.Validate() != nil || authority.ScopeKind != jobsv1.ScopeTenant ||
		authority.WriterOrigin != jobsv1.WriterV2 || scopeID != scope.TenantID() {
		return authorityStamp{}, ErrAuthorityMismatch
	}
	return authorityStamp{scopeID: scopeID, groupID: authority.GroupID,
		groupGeneration: authority.GroupGeneration, domain: authority.Domain,
		generation: authority.CutoverGeneration}, nil
}
