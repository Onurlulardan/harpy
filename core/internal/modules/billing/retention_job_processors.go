package billing

import (
	"fmt"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	"github.com/google/uuid"
)

func NewRetentionJobProcessors(
	release platformasync.ReleaseCatalogs,
	store jobexecution.DatabaseJobStore,
) ([]jobexecution.Processor, error) {
	jobs, _, err := release.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("read billing retention job catalog: %w", err)
	}
	runtime, err := NewRetentionJobRuntime(uuid.NewV7)
	if err != nil {
		return nil, fmt.Errorf("build billing retention runtime: %w", err)
	}
	return runtime.Processors(jobs, store)
}

func (runtime *RetentionJobRuntime) Processors(
	catalog jobsv1.Catalog,
	store jobexecution.DatabaseJobStore,
) ([]jobexecution.Processor, error) {
	if runtime == nil {
		return nil, fmt.Errorf("billing retention job runtime is nil")
	}
	policy, err := catalog.Resolve(asyncmanifestv1.BillingRetentionJobKind, 1)
	if err != nil || policy.OwnerDomain != "billing" ||
		policy.EffectClass != jobsv1.EffectDatabaseTransactional ||
		policy.MaximumAttempts != jobsv1.DefaultMaximumAttempts {
		return nil, fmt.Errorf("billing retention job policy is invalid")
	}
	processor, err := jobexecution.NewDatabaseProcessor(policy, store, runtime.binding)
	if err != nil {
		return nil, fmt.Errorf("build billing retention job processor: %w", err)
	}
	return []jobexecution.Processor{processor}, nil
}
