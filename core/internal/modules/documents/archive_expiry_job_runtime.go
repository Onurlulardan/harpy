package documents

import (
	"context"
	"fmt"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	archivepostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/archive"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type ArchiveExpiryJobRuntime struct {
	binding platformasync.DatabaseJobBinding
}

func NewArchiveExpiryJobRuntime(ids foundation.IDSource,
	clock foundation.Clock,
) (*ArchiveExpiryJobRuntime, error) {
	command, err := archiveapp.NewExpiryCommand(ids, clock)
	if err != nil {
		return nil, err
	}
	binding, err := platformasync.NewRepeatableReadDatabaseJobBinding("documents",
		func(database platformtx.JobBoundExecutor) platformasync.DatabaseJobCommand {
			return &archiveExpiryDatabaseJob{database: database, command: command,
				clock: clock}
		})
	if err != nil {
		return nil, fmt.Errorf("build Documents archive expiry binding: %w", err)
	}
	return &ArchiveExpiryJobRuntime{binding: binding}, nil
}

func (runtime *ArchiveExpiryJobRuntime) Binding() platformasync.DatabaseJobBinding {
	if runtime == nil {
		return platformasync.DatabaseJobBinding{}
	}
	return runtime.binding
}

func (runtime *ArchiveExpiryJobRuntime) Processor(catalog jobsv1.Catalog,
	store jobexecution.DatabaseJobStore,
) (jobexecution.Processor, error) {
	if runtime == nil {
		return nil, archiveapp.ErrInvalid
	}
	policy, err := catalog.Resolve(asyncmanifestv1.DocumentsArchiveExpireJobKind, 1)
	if err != nil || policy.OwnerDomain != "documents" ||
		policy.EffectClass != jobsv1.EffectDatabaseTransactional ||
		policy.MaximumAttempts != jobsv1.DefaultMaximumAttempts {
		return nil, fmt.Errorf("Documents archive expiry policy is invalid")
	}
	return jobexecution.NewDatabaseProcessor(policy, store, runtime.binding)
}

type archiveExpiryDatabaseJob struct {
	database platformtx.JobBoundExecutor
	command  *archiveapp.ExpiryCommand
	clock    foundation.Clock
}

func (job *archiveExpiryDatabaseJob) Execute(ctx context.Context,
	claim platformasync.DatabaseJobClaim,
) (platformasync.DatabaseJobOutcome, error) {
	parsed, tenantID, err := parseArchiveExpiryClaim(claim)
	if job == nil || job.command == nil || job.clock == nil || err != nil {
		return platformasync.DatabaseJobOutcome{}, archiveapp.ErrInvalid
	}
	store, err := archivepostgres.NewJobStore(job.database, tenantID)
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	execution, err := job.command.Expire(ctx, store, foundation.ID(parsed.ResourceID))
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	return archiveExpiryOutcome(tenantID.String(), execution, job.clock.Now())
}
