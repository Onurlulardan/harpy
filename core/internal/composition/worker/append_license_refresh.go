package worker

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func appendLicenseRefreshWorkload(
	ctx context.Context,
	pool *postgres.Pool,
	settings config.Worker,
	version string,
	workloads []managedAsyncWorkload,
) ([]managedAsyncWorkload, error) {
	workload, enabled, err := buildLicenseRefreshWorkload(
		ctx, pool, settings, version,
	)
	if err != nil {
		return nil, fmt.Errorf("build license refresh workload: %w", err)
	}
	if enabled {
		workloads = append(workloads, workload)
	}
	return workloads, nil
}
