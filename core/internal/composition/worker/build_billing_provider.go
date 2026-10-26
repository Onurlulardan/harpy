package worker

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

func buildBillingProviderProcessors(
	release platformasync.ReleaseCatalogs,
	store asyncpostgres.WorkerStore,
) ([]jobexecution.Processor, error) {
	processors, err := billing.NewManualProviderJobProcessors(release, store)
	if err != nil {
		return nil, fmt.Errorf("build billing provider processors: %w", err)
	}
	return processors, nil
}
