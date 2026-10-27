package work

import (
	"fmt"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

func NewWorkJobProcessors(release platformasync.ReleaseCatalogs,
	store jobexecution.DatabaseJobStore,
) ([]jobexecution.Processor, error) {
	jobs, events, err := release.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("read Work async catalogs: %w", err)
	}
	producer, err := events.Producers().Resolve(asyncmanifestv1.WorkReminderDueEventType, 1)
	if err != nil || producer.OwnerDomain != "work" ||
		producer.PayloadSchemaSHA256 != asyncmanifestv1.WorkReminderDueSchemaSHA256 {
		return nil, fmt.Errorf("Work reminder producer policy is invalid")
	}
	runtime, err := NewWorkJobRuntime()
	if err != nil {
		return nil, err
	}
	return runtime.Processors(jobs, store)
}

func (runtime *WorkJobRuntime) Processors(catalog jobsv1.Catalog,
	store jobexecution.DatabaseJobStore,
) ([]jobexecution.Processor, error) {
	if runtime == nil {
		return nil, fmt.Errorf("Work job runtime is nil")
	}
	policies := asyncmanifestv1.WorkJobPolicies()
	processors := make([]jobexecution.Processor, 0, len(policies))
	for _, expected := range policies {
		policy, err := catalog.Resolve(expected.Kind, expected.PayloadRevision)
		if err != nil || policy != expected || policy.OwnerDomain != "work" ||
			policy.EffectClass != jobsv1.EffectDatabaseTransactional ||
			policy.MaximumAttempts != jobsv1.DefaultMaximumAttempts {
			return nil, fmt.Errorf("Work job policy %s is invalid", expected.Kind)
		}
		processor, err := jobexecution.NewDatabaseProcessor(policy, store, runtime.binding)
		if err != nil {
			return nil, fmt.Errorf("build Work job processor %s: %w", policy.Kind, err)
		}
		processors = append(processors, processor)
	}
	return processors, nil
}
