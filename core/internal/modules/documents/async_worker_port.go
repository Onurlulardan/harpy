package documents

import (
	"context"
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type JobActivity interface {
	CheckJobActive(context.Context, uuid.UUID, jobsv1.Authority) error
}

type WorkerAsyncPort struct {
	catalogs platformasync.WriteCatalogs
	jobs     jobsv1.Catalog
	activity JobActivity
	clock    foundation.Clock
}

func NewWorkerAsyncPort(catalogs platformasync.WriteCatalogs,
	activity JobActivity, clock foundation.Clock,
) (*WorkerAsyncPort, error) {
	jobs, _, err := catalogs.Snapshot()
	if err != nil || activity == nil || clock == nil {
		return nil, fmt.Errorf("Documents worker async dependencies are required")
	}
	return &WorkerAsyncPort{catalogs: catalogs, jobs: jobs,
		activity: activity, clock: clock}, nil
}

func (port *WorkerAsyncPort) Check(ctx context.Context, jobID foundation.ID) error {
	authority, found := jobAuthorityFromContext(ctx)
	if port == nil || !found || !jobID.UUIDv7() ||
		port.activity.CheckJobActive(ctx, uuid.UUID(jobID), authority) != nil {
		return ErrDocumentsRuntime
	}
	return nil
}

func workerScopeAuthority(ctx context.Context,
	scope platformtx.JobTenantScope,
) (jobsv1.Authority, bool) {
	authority, found := jobAuthorityFromContext(ctx)
	if !found || scope.TenantID().String() != authority.ScopeID {
		return jobsv1.Authority{}, false
	}
	return authority, true
}
