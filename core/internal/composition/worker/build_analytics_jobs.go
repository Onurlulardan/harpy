package worker

import (
	"fmt"

	analyticsmodule "github.com/Onurlulardan/harpy/core/internal/modules/analytics"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

func buildAnalyticsJobProcessors(
	release platformasync.ReleaseCatalogs,
	store jobexecution.DatabaseJobStore,
) ([]jobexecution.Processor, error) {
	processors, err := analyticsmodule.NewJobProcessors(release, store)
	if err != nil {
		return nil, fmt.Errorf("build Analytics job processors: %w", err)
	}
	return processors, nil
}
