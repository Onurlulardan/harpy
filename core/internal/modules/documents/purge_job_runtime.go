package documents

import (
	"context"
	"fmt"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	purgepostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/purge"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	purgeapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/purge"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type PurgeJobRuntime struct {
	binding platformasync.DatabaseJobBinding
}

func NewPurgeJobRuntime(ids foundation.IDSource,
	clock foundation.Clock,
) (*PurgeJobRuntime, error) {
	command, err := purgeapp.NewCommand(ids, clock)
	if err != nil {
		return nil, err
	}
	binding, err := platformasync.NewRepeatableReadDatabaseJobBinding("documents",
		func(database platformtx.JobBoundExecutor) platformasync.DatabaseJobCommand {
			return &purgeDatabaseJobCommand{database: database, command: command,
				ids: ids, clock: clock}
		})
	if err != nil {
		return nil, fmt.Errorf("build Documents purge binding: %w", err)
	}
	return &PurgeJobRuntime{binding: binding}, nil
}

func (runtime *PurgeJobRuntime) Binding() platformasync.DatabaseJobBinding {
	if runtime == nil {
		return platformasync.DatabaseJobBinding{}
	}
	return runtime.binding
}

func (runtime *PurgeJobRuntime) Processor(catalog jobsv1.Catalog,
	store jobexecution.DatabaseJobStore,
) (jobexecution.Processor, error) {
	if runtime == nil {
		return nil, purgeapp.ErrInvalid
	}
	policy, err := catalog.Resolve(asyncmanifestv1.DocumentsItemPurgeJobKind, 1)
	if err != nil || policy.OwnerDomain != "documents" ||
		policy.EffectClass != jobsv1.EffectDatabaseTransactional ||
		policy.MaximumAttempts != jobsv1.DefaultMaximumAttempts {
		return nil, fmt.Errorf("Documents purge policy is invalid")
	}
	processor, err := jobexecution.NewDatabaseProcessor(policy, store, runtime.binding)
	if err != nil {
		return nil, fmt.Errorf("build Documents purge processor: %w", err)
	}
	return processor, nil
}

type purgeDatabaseJobCommand struct {
	database platformtx.JobBoundExecutor
	command  *purgeapp.Command
	ids      foundation.IDSource
	clock    foundation.Clock
}

var _ platformasync.DatabaseJobCommand = (*purgeDatabaseJobCommand)(nil)

func (command *purgeDatabaseJobCommand) Execute(ctx context.Context,
	claim platformasync.DatabaseJobClaim,
) (platformasync.DatabaseJobOutcome, error) {
	parsed, tenantID, err := parsePurgeClaim(claim)
	if command == nil || command.command == nil || command.ids == nil ||
		command.clock == nil || err != nil {
		return platformasync.DatabaseJobOutcome{}, purgeapp.ErrInvalid
	}
	store, err := purgepostgres.NewJobStore(command.database, tenantID)
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	execution, err := command.command.Resume(ctx, store, foundation.ID(parsed.ResourceID),
		foundation.ID(parsed.AfterDeletionID), maximumPurgeChildren)
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	return command.outcome(parsed, execution)
}
