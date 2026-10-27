package worker

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

func buildBillingSubscriptionProcessors(
	release platformasync.ReleaseCatalogs,
	store jobexecution.DatabaseJobStore,
) ([]jobexecution.Processor, error) {
	processors, err := billing.NewSubscriptionJobProcessors(release, store)
	if err != nil {
		return nil, fmt.Errorf("build billing subscription processors: %w", err)
	}
	return processors, nil
}
