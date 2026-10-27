package worker

import (
	"fmt"

	workmodule "github.com/Onurlulardan/harpy/core/internal/modules/work"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

func buildWorkJobProcessors(release platformasync.ReleaseCatalogs,
	store asyncpostgres.WorkerStore,
) ([]jobexecution.Processor, error) {
	processors, err := workmodule.NewWorkJobProcessors(release, store)
	if err != nil {
		return nil, fmt.Errorf("build Work job processors: %w", err)
	}
	return processors, nil
}
