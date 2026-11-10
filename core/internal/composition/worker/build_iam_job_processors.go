package worker

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func buildIAMJobProcessors(
	ctx context.Context,
	pool *platformpostgres.Pool,
	workerSettings config.Worker,
	settings config.IAMWorker,
	coordinator *platformtx.Coordinator,
	release platformasync.ReleaseCatalogs,
	writeCatalogs platformasync.WriteCatalogs,
	store asyncpostgres.WorkerStore,
	gate *admission.Gate,
) ([]jobexecution.Processor, *iamWorkerResources, error) {
	jobCatalog, _, err := release.Snapshot()
	if err != nil {
		return nil, nil, fmt.Errorf("read IAM worker job catalog: %w", err)
	}
	maintenance, err := buildIAMMaintenance(settings)
	if err != nil {
		return nil, nil, err
	}
	delivery, lifecycleDelivery, err := buildIAMDelivery(ctx, settings, coordinator)
	if err != nil {
		_ = maintenance.Close()
		return nil, nil, err
	}
	resources, err := newIAMWorkerResources(maintenance, delivery, lifecycleDelivery)
	if err != nil {
		_ = lifecycleDelivery.Close()
		_ = delivery.Close()
		_ = maintenance.Close()
		return nil, nil, err
	}
	owned := true
	defer func() {
		if owned {
			_ = resources.Close()
		}
	}()
	processors, err := maintenance.Processors(jobCatalog, store)
	if err != nil {
		return nil, nil, fmt.Errorf("build IAM maintenance processors: %w", err)
	}
	deliveryProcessors, err := delivery.Processors(jobCatalog, store)
	if err != nil {
		return nil, nil, fmt.Errorf("build IAM invitation delivery processor: %w", err)
	}
	processors = append(processors, deliveryProcessors...)
	lifecycleProcessors, err := lifecycleDelivery.Processors(jobCatalog, store)
	if err != nil {
		return nil, nil, fmt.Errorf("build IAM lifecycle delivery processor: %w", err)
	}
	processors = append(processors, lifecycleProcessors...)
	owned = false
	return processors, resources, nil
}
