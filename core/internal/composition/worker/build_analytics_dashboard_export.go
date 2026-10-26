package worker

import (
	"context"
	"fmt"
	"time"

	analyticsmodule "github.com/Onurlulardan/harpy/core/internal/modules/analytics"
	billingmodule "github.com/Onurlulardan/harpy/core/internal/modules/billing"
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type analyticsDashboardExportRuntime struct {
	workload   managedAsyncWorkload
	processors []jobexecution.Processor
}

func buildAnalyticsDashboardExportRuntime(
	ctx context.Context,
	settings config.Worker,
	release platformasync.ReleaseCatalogs,
	catalogs platformasync.WriteCatalogs,
	jobStore asyncpostgres.WorkerStore,
	pool *platformpostgres.Pool,
	coordinator *platformtx.Coordinator,
	gate *admission.Gate,
) (analyticsDashboardExportRuntime, error) {
	if ctx == nil || ctx.Err() != nil || jobStore == nil || pool == nil ||
		coordinator == nil || gate == nil {
		return analyticsDashboardExportRuntime{},
			fmt.Errorf("Analytics dashboard export worker dependencies are incomplete")
	}
	store, err := analyticsmodule.BindDashboardExportWorkerStore(pool, coordinator)
	if err != nil {
		return analyticsDashboardExportRuntime{}, err
	}
	iam, err := iammodule.NewAnalyticsWorkerAuthorizationRuntime()
	if err != nil {
		return analyticsDashboardExportRuntime{}, err
	}
	authorizer := analyticsDashboardExportAuthorizer{
		pool: pool, coordinator: coordinator, iam: iam,
	}
	executor, err := buildAnalyticsDashboardExportQueryExecutor(
		catalogs, pool, coordinator,
	)
	if err != nil {
		return analyticsDashboardExportRuntime{}, err
	}
	renderer, err := analyticsmodule.NewDashboardExportRenderer(executor)
	if err != nil {
		return analyticsDashboardExportRuntime{},
			fmt.Errorf("build Analytics dashboard export renderer: %w", err)
	}
	artifacts, err := openAnalyticsDashboardExportArtifacts(ctx, settings)
	if err != nil {
		return analyticsDashboardExportRuntime{}, err
	}
	values := analyticsDashboardExportValues{}
	workload, worker, err := analyticsmodule.NewDashboardExportWorkerWorkload(
		values, values, store, authorizer, renderer, artifacts, gate,
	)
	if err != nil {
		return analyticsDashboardExportRuntime{}, err
	}
	processors, err := analyticsmodule.NewDashboardExportJobProcessors(
		release, jobStore, worker,
	)
	if err != nil {
		return analyticsDashboardExportRuntime{},
			fmt.Errorf("build Analytics dashboard export processors: %w", err)
	}
	return analyticsDashboardExportRuntime{
		workload: workload, processors: processors,
	}, nil
}

func buildAnalyticsDashboardExportQueryExecutor(
	catalogs platformasync.WriteCatalogs,
	pool *platformpostgres.Pool,
	coordinator *platformtx.Coordinator,
) (analyticsmodule.DashboardExportWidgetExecutor, error) {
	entitlements, err := billingmodule.NewMeteredResourceEntitlements(
		catalogs, uuid.NewV7, time.Now,
	)
	if err != nil {
		return nil, fmt.Errorf("build Analytics worker entitlements: %w", err)
	}
	return analyticsmodule.NewDashboardExportQueryExecutor(
		analyticsmodule.PlatformQueryRuntimeDependencies{
			Generations: pool, Scopes: coordinator, Entitlements: entitlements,
			Catalogs: catalogs,
		})
}
