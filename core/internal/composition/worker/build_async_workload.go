package worker

import (
	"context"
	"fmt"

	analyticsmodule "github.com/Onurlulardan/harpy/core/internal/modules/analytics"
	iam "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	"github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/eventdelivery"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/outboxdelivery"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/retention"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func buildAsyncWorkload(
	ctx context.Context,
	pool *postgres.Pool,
	workerSettings config.Worker,
	settings config.RabbitMQ,
	iamSettings config.IAMWorker,
	retentionObserver retention.Observer,
	gate *admission.Gate,
	version string,
) (*asyncSupervisor, *iamWorkerResources, error) {
	if gate == nil {
		return nil, nil, fmt.Errorf("worker admission gate is required")
	}
	release, err := async.LoadReleaseCatalogs()
	if err != nil {
		return nil, nil, fmt.Errorf("build canonical async manifest: %w", err)
	}
	jobCatalog, catalogs, err := release.Snapshot()
	if err != nil {
		return nil, nil, fmt.Errorf("read canonical async manifest: %w", err)
	}
	writeCatalogs, err := async.LoadWriteCatalogs()
	if err != nil {
		return nil, nil, fmt.Errorf("build async write catalogs: %w", err)
	}
	store, err := asyncpostgres.BindPublisherStore(pool)
	if err != nil {
		return nil, nil, fmt.Errorf("bind outbox publisher store: %w", err)
	}
	if err := store.AssertOutboxCatalog(ctx, catalogs.Producers()); err != nil {
		return nil, nil, fmt.Errorf("assert live outbox producer compatibility: %w", err)
	}
	consumers := catalogs.Consumers()
	connector, err := outboxdelivery.NewRabbitConnector(settings, consumers)
	if err != nil {
		return nil, nil, fmt.Errorf("build RabbitMQ connector: %w", err)
	}
	if err := probeRabbitMQ(ctx, connector); err != nil {
		return nil, nil, err
	}
	outboxRunner, err := outboxdelivery.New(
		store, catalogs.Producers(), connector, outboxdelivery.Settings{
			ConfirmTimeout:   settings.ConfirmTimeout(),
			ReconnectMinimum: settings.ReconnectMinimum(),
			ReconnectMaximum: settings.ReconnectMaximum(),
		},
		gate,
	)
	if err != nil {
		return nil, nil, err
	}
	coordinator, err := platformtx.NewCoordinator(pool)
	if err != nil {
		return nil, nil, fmt.Errorf("build worker tenant coordinator: %w", err)
	}
	jobStore, err := asyncpostgres.BindWorker(pool, coordinator)
	if err != nil {
		return nil, nil, fmt.Errorf("bind job worker store: %w", err)
	}
	exportRecovery, err := iam.OpenSystemTransferExportReconciliationRecovery(coordinator)
	if err != nil {
		return nil, nil, fmt.Errorf("bind IAM export reconciliation recovery: %w", err)
	}
	processors, err := buildCoreJobProcessors(
		workerSettings, release, writeCatalogs, jobStore, pool, coordinator,
	)
	if err != nil {
		return nil, nil, err
	}
	analyticsExport, err := buildAnalyticsDashboardExportRuntime(
		ctx, workerSettings, release, writeCatalogs, jobStore, pool, coordinator, gate,
	)
	if err != nil {
		return nil, nil, err
	}
	processors = append(processors, analyticsExport.processors...)
	iamProcessors, iamResources, err := buildIAMJobProcessors(
		ctx, pool, workerSettings, iamSettings, coordinator, release,
		writeCatalogs, jobStore, gate,
	)
	if err != nil {
		return nil, nil, err
	}
	processors = append(processors, iamProcessors...)
	resourcesOwned := true
	defer func() {
		if resourcesOwned {
			_ = iamResources.Close()
		}
	}()
	registry, err := analyticsmodule.NewDashboardExportWorkerJobRegistry(
		jobCatalog, processors, nil,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("build job processor registry: %w", err)
	}
	if err := jobStore.AssertJobCatalog(ctx, jobCatalog); err != nil {
		return nil, nil, fmt.Errorf("assert live job processor compatibility: %w", err)
	}
	runnerStore, err := newIAMExportReconciliationRecoveryStore(jobStore, exportRecovery)
	if err != nil {
		return nil, nil, err
	}
	jobRunner, err := jobexecution.NewRunner(runnerStore, registry, gate)
	if err != nil {
		return nil, nil, fmt.Errorf("build job runner: %w", err)
	}
	eventRegistry, err := eventdelivery.NewRegistry(catalogs.Consumers(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("build event handler registry: %w", err)
	}
	eventRunner, err := eventdelivery.New(
		settings, catalogs.Producers(), eventRegistry, gate,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("build event delivery runner: %w", err)
	}
	retentionRunner, err := buildAsyncRetention(pool, retentionObserver, gate)
	if err != nil {
		return nil, nil, err
	}
	supervisor, err := assembleWorkerAsyncSupervisor(
		ctx, pool, workerSettings, iamSettings, settings, catalogs, gate, version,
		[]managedAsyncWorkload{
			outboxRunner, jobRunner, eventRunner, retentionRunner,
			analyticsExport.workload, iamResources.cleanup,
		},
	)
	if err != nil {
		return nil, nil, err
	}
	resourcesOwned = false
	return supervisor, iamResources, nil
}
