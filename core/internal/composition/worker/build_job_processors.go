package worker

import (
	"github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func buildCoreJobProcessors(
	settings config.Worker,
	release async.ReleaseCatalogs,
	catalogs async.WriteCatalogs,
	store asyncpostgres.WorkerStore,
	pool *postgres.Pool,
	coordinator *platformtx.Coordinator,
) ([]jobexecution.Processor, error) {
	processors, err := buildEventGapProcessors(release, store)
	if err != nil {
		return nil, err
	}
	builders := []func(async.ReleaseCatalogs, asyncpostgres.WorkerStore) (
		[]jobexecution.Processor, error,
	){
		func(release async.ReleaseCatalogs, store asyncpostgres.WorkerStore) (
			[]jobexecution.Processor, error,
		) {
			return buildBillingSubscriptionProcessors(release, store)
		},
		buildBillingProviderProcessors,
		func(release async.ReleaseCatalogs, store asyncpostgres.WorkerStore) (
			[]jobexecution.Processor, error,
		) {
			return buildBillingRetentionProcessors(release, store)
		},
		func(release async.ReleaseCatalogs, store asyncpostgres.WorkerStore) (
			[]jobexecution.Processor, error,
		) {
			return buildNotificationsRetentionProcessors(release, store)
		},
		buildWorkJobProcessors,
	}
	for _, build := range builders {
		built, err := build(release, store)
		if err != nil {
			return nil, err
		}
		processors = append(processors, built...)
	}
	deliveryProcessors, err := buildNotificationsDeliveryProcessors(
		settings, release, store, coordinator)
	if err != nil {
		return nil, err
	}
	processors = append(processors, deliveryProcessors...)
	analyticsProcessors, err := buildAnalyticsJobProcessors(release, store)
	if err != nil {
		return nil, err
	}
	processors = append(processors, analyticsProcessors...)
	catalogProcessor, err := buildCatalogJobProcessor(
		settings, release, catalogs, pool, coordinator,
	)
	if err != nil {
		return nil, err
	}
	processors = append(processors, catalogProcessor)
	return processors, nil
}
