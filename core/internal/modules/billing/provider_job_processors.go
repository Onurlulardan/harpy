package billing

import (
	"fmt"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

func NewManualProviderJobProcessors(
	release platformasync.ReleaseCatalogs,
	store asyncpostgres.WorkerStore,
) ([]jobexecution.Processor, error) {
	jobs, _, err := release.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("read billing provider job catalog: %w", err)
	}
	registry, err := NewManualProviderRegistry()
	if err != nil {
		return nil, fmt.Errorf("build manual billing provider registry: %w", err)
	}
	runtime, err := NewProviderJobRuntime(registry)
	if err != nil {
		return nil, fmt.Errorf("build billing provider job runtime: %w", err)
	}
	processors, err := runtime.Processors(jobs, store)
	if err != nil {
		return nil, err
	}
	webhook, err := newManualOnlyWebhookJobProcessor(jobs, store)
	if err != nil {
		return nil, err
	}
	return append([]jobexecution.Processor{webhook}, processors...), nil
}
