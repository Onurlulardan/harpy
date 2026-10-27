package worker

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func assembleWorkerAsyncSupervisor(
	ctx context.Context,
	pool *postgres.Pool,
	workerSettings config.Worker,
	iamSettings config.IAMWorker,
	broker config.RabbitMQ,
	catalogs platformasync.EventCatalogs,
	gate *admission.Gate,
	version string,
	workloads []managedAsyncWorkload,
) (*asyncSupervisor, error) {
	customerProvisioning, enabled, err := buildCustomerProvisioningWorkload(
		ctx, pool, workerSettings, iamSettings,
	)
	if err != nil {
		return nil, fmt.Errorf("build customer provisioning workload: %w", err)
	}
	if enabled {
		workloads = append(workloads, customerProvisioning)
	}
	workloads, err = appendLicenseRefreshWorkload(
		ctx, pool, workerSettings, version, workloads,
	)
	if err != nil {
		return nil, err
	}
	workloads, err = appendAuthorityDeliveryWorkload(ctx, pool, workerSettings, workloads)
	if err != nil {
		return nil, err
	}
	deadLetters, deadLettersEnabled, err := buildDeadLetterWorkload(
		ctx, pool, broker, catalogs, gate,
	)
	if err != nil {
		return nil, err
	}
	if deadLettersEnabled {
		workloads = append(workloads, deadLetters)
	}
	return newAsyncSupervisor(workloads...)
}
