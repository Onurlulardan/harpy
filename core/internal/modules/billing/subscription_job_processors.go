package billing

import (
	"fmt"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	"github.com/google/uuid"
)

func NewSubscriptionJobProcessors(
	release platformasync.ReleaseCatalogs,
	store jobexecution.DatabaseJobStore,
) ([]jobexecution.Processor, error) {
	jobs, _, err := release.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("read billing subscription job catalog: %w", err)
	}
	runtime, err := NewSubscriptionJobRuntime(
		uuid.NewV7,
		func() time.Time { return time.Now().UTC().Truncate(time.Microsecond) },
	)
	if err != nil {
		return nil, fmt.Errorf("build billing subscription runtime: %w", err)
	}
	return runtime.Processors(jobs, store)
}

func (runtime *SubscriptionJobRuntime) Processors(
	catalog jobsv1.Catalog,
	store jobexecution.DatabaseJobStore,
) ([]jobexecution.Processor, error) {
	if runtime == nil {
		return nil, fmt.Errorf("billing subscription job runtime is nil")
	}
	policy, err := catalog.Resolve(subscriptionBoundaryJobKind, 1)
	if err != nil || policy.OwnerDomain != "billing" ||
		policy.EffectClass != jobsv1.EffectDatabaseTransactional ||
		policy.MaximumAttempts != jobsv1.DefaultMaximumAttempts {
		return nil, fmt.Errorf("billing subscription job policy is invalid")
	}
	processor, err := jobexecution.NewDatabaseProcessor(
		policy, store, runtime.binding,
	)
	if err != nil {
		return nil, fmt.Errorf("build billing subscription job processor: %w", err)
	}
	return []jobexecution.Processor{processor}, nil
}
