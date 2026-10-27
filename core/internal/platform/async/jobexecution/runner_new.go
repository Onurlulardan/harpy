package jobexecution

import (
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	"github.com/google/uuid"
)

func NewRunner(store Store, registry Registry, gate *admission.Gate) (*Runner, error) {
	if store == nil || gate == nil {
		return nil, ErrRunnerDependencies
	}
	validated, err := NewRegistry(registryProcessors(registry))
	if err != nil {
		return nil, fmt.Errorf("validate job processor registry: %w", err)
	}
	return &Runner{
		store: store, registry: validated, workerID: uuid.New(),
		admission: gate, lifecycle: newRunnerLifecycle(),
	}, nil
}

func registryProcessors(registry Registry) []Processor {
	policies := registry.Catalog().Policies()
	processors := make([]Processor, 0, len(policies))
	for _, policy := range policies {
		registered, exists := registry.processors[processorKey(policy.Kind, policy.PayloadRevision)]
		if exists {
			processors = append(processors, registered.processor)
		}
	}
	return processors
}

func (runner *Runner) Catalog() jobsv1.Catalog { return runner.registry.Catalog() }
