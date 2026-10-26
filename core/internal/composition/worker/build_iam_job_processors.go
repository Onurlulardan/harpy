package worker

import (
	"context"
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
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
	artifacts, err := buildIAMSystemTransferArtifacts(ctx, workerSettings)
	if err != nil {
		return nil, nil, err
	}
	cleanup, err := buildIAMImportSourceCleanup(pool, coordinator, artifacts, gate)
	if err != nil {
		return nil, nil, err
	}
	cursorKeys, err := workerSettings.RestQueryCursorKeyring()
	if err != nil {
		return nil, nil, fmt.Errorf("load IAM import cursor keys: %w", err)
	}
	imports, err := iammodule.OpenSystemTransferImportJobRuntime(
		ctx, pool, coordinator, release, writeCatalogs, cursorKeys, settings, artifacts,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("build IAM system import processor: %w", err)
	}
	exports, err := iammodule.OpenSystemTransferExportJobRuntime(
		ctx, pool, coordinator, release, cursorKeys, store, artifacts,
	)
	if err != nil {
		_ = imports.Close()
		return nil, nil, fmt.Errorf("build IAM system export processors: %w", err)
	}
	maintenance, err := buildIAMMaintenance(settings)
	if err != nil {
		_ = exports.Close()
		_ = imports.Close()
		return nil, nil, err
	}
	delivery, lifecycleDelivery, err := buildIAMDelivery(ctx, settings, coordinator)
	if err != nil {
		_ = maintenance.Close()
		_ = imports.Close()
		_ = exports.Close()
		return nil, nil, err
	}
	resources, err := newIAMWorkerResources(
		maintenance, delivery, lifecycleDelivery, imports, exports, cleanup,
	)
	if err != nil {
		_ = lifecycleDelivery.Close()
		_ = delivery.Close()
		_ = maintenance.Close()
		_ = imports.Close()
		_ = exports.Close()
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
	importProcessor, err := imports.Processor()
	if err != nil {
		return nil, nil, fmt.Errorf("open IAM system import processor: %w", err)
	}
	exportProcessors, err := exports.Processors()
	if err != nil {
		return nil, nil, fmt.Errorf("open IAM system export processors: %w", err)
	}
	processors, err = appendIAMTransferProcessors(
		processors, importProcessor, exportProcessors,
	)
	if err != nil {
		return nil, nil, err
	}
	owned = false
	return processors, resources, nil
}
