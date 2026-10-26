package postgres

import (
	"context"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type workerDatabase interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}

type workerIdentity struct{ marker byte }

type workerRepository struct {
	database          workerDatabase
	coordinator       *platformtx.Coordinator
	databaseJobBinder platformtx.JobOwnerBinder[*boundDatabaseJobStore]
	cleanup           time.Duration
	identity          *workerIdentity
}

type WorkerStore interface {
	JobLifecycleStore
	DatabaseJobStore
	CheckJobActive(context.Context, uuid.UUID, jobsv1.Authority) error
	BeginJobAuthorityPage(int) (JobAuthorityPageToken, error)
	ListJobAuthorities(context.Context, JobAuthorityPageToken) (JobAuthorityPage, error)
	ClaimJobs(
		context.Context, JobAuthority, uuid.UUID, int, jobsv1.Catalog,
	) ([]JobClaim, error)
	ReconcileEventGap(context.Context, JobClaim) (EventGapReconciliation, error)
	ResolveStaleJobAuthority(context.Context, uuid.UUID) (JobAuthority, error)
	AssertJobCatalog(context.Context, jobsv1.Catalog) error
}

type JobLifecycleStore interface {
	HeartbeatJob(context.Context, JobClaim) (JobLease, error)
	ResolveJobControl(context.Context, JobAuthority, uuid.UUID) (JobControl, error)
	RequestJobCancellation(context.Context, JobControl, JobCancellation) (JobCancelResult, error)
	CancelJob(context.Context, JobClaim, JobCancellation) (time.Time, error)
	CompleteExternalJob(
		context.Context, ExternalEffectPermit, effectsv1.Receipt,
	) (time.Time, error)
	MarkExternalJobOutcomeUnknown(context.Context, ExternalEffectPermit) (time.Time, error)
	FailExternalJob(
		context.Context, ExternalEffectPermit, effectsv1.Receipt, JobFailure,
	) (JobFailureResult, error)
	RecoverExpiredJobs(context.Context, JobAuthority, int) (JobRecoveryResult, error)
	ReconcileStaleJobs(context.Context, int) (int, error)
	DisposeStaleJob(
		context.Context, JobAuthority, StaleJobDispositionCommand,
	) (StaleJobDispositionResult, error)
	ParkStaleJob(context.Context, JobClaim) (bool, error)
}

func (repository *workerRepository) valid() bool {
	return repository != nil && repository.database != nil &&
		repository.cleanup > 0 && repository.identity != nil &&
		repository.identity.marker == 1
}
