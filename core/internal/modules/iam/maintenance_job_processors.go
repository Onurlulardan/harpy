package iam

import (
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

func (runtime *MaintenanceJobRuntime) Processors(
	catalog jobsv1.Catalog,
	store jobexecution.DatabaseJobStore,
) ([]jobexecution.Processor, error) {
	if runtime == nil || runtime.closed.Load() {
		return nil, fmt.Errorf("IAM maintenance runtime is closed")
	}
	keys := runtime.JobKeys()
	processors := make([]jobexecution.Processor, 0, len(keys))
	for _, key := range keys {
		policy, err := catalog.Resolve(key.Kind, key.PayloadRevision)
		if err != nil || policy.OwnerDomain != "iam" ||
			policy.EffectClass != jobsv1.EffectDatabaseTransactional ||
			policy.MaximumAttempts != jobsv1.DefaultMaximumAttempts {
			return nil, fmt.Errorf("IAM maintenance job policy %s/%d is invalid",
				key.Kind, key.PayloadRevision)
		}
		processor, err := jobexecution.NewDatabaseProcessor(
			policy, store, runtime.binding,
		)
		if err != nil {
			return nil, fmt.Errorf("build IAM maintenance processor: %w", err)
		}
		processors = append(processors, processor)
	}
	return processors, nil
}
