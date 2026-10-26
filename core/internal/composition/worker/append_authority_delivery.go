package worker

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func appendAuthorityDeliveryWorkload(
	ctx context.Context,
	pool *postgres.Pool,
	settings config.Worker,
	workloads []managedAsyncWorkload,
) ([]managedAsyncWorkload, error) {
	if settings.Authority().Mode() == config.WorkerAuthorityConnected {
		return workloads, nil
	}
	workload, err := buildAuthorityDeliveryWorkload(ctx, pool, settings)
	if err != nil {
		return nil, fmt.Errorf("build platform authority delivery workload: %w", err)
	}
	return append(workloads, workload), nil
}
