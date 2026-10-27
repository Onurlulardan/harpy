package worker

import (
	"fmt"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

func buildEventGapProcessors(
	release platformasync.ReleaseCatalogs,
	store jobexecution.EventGapReconciliationStore,
) ([]jobexecution.Processor, error) {
	policies, err := release.GapJobs()
	if err != nil {
		return nil, fmt.Errorf("read event gap processor catalog: %w", err)
	}
	processors := make([]jobexecution.Processor, 0, len(policies))
	for _, policy := range policies {
		processor, err := jobexecution.NewEventGapProcessor(policy, store)
		if err != nil {
			return nil, fmt.Errorf("build %s processor: %w", policy.Kind, err)
		}
		processors = append(processors, processor)
	}
	return processors, nil
}
