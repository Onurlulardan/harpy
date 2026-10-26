package worker

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/retention"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func buildAsyncRetention(
	pool *platformpostgres.Pool,
	observer retention.Observer,
	gate *admission.Gate,
) (*retention.Runner, error) {
	store, err := asyncpostgres.BindRetention(pool)
	if err != nil {
		return nil, fmt.Errorf("bind async retention workload: %w", err)
	}
	runner, err := retention.NewRunner(store, observer, gate)
	if err != nil {
		return nil, fmt.Errorf("build async retention workload: %w", err)
	}
	return runner, nil
}
