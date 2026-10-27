package analytics

import (
	"context"
	"fmt"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	"github.com/google/uuid"
)

type dashboardExportJobWorker interface {
	ProcessDelete(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) (
		exportapp.RenderResult, error)
	ProcessExpiry(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) (
		exportapp.RenderResult, error)
}

func NewDashboardExportJobProcessors(
	release platformasync.ReleaseCatalogs,
	store asyncpostgres.WorkerStore,
	worker *DashboardExportWorker,
) ([]jobexecution.Processor, error) {
	if store == nil || worker == nil {
		return nil, exportapp.ErrUnavailable
	}
	jobs, _, err := release.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("read Analytics dashboard export job catalog: %w", err)
	}
	return newDashboardExportJobProcessors(jobs, store, worker, time.Now)
}

func newDashboardExportJobProcessors(
	jobs jobsv1.Catalog,
	store asyncpostgres.WorkerStore,
	worker dashboardExportJobWorker,
	now func() time.Time,
) ([]jobexecution.Processor, error) {
	if store == nil || worker == nil || now == nil {
		return nil, exportapp.ErrUnavailable
	}
	preEffect, err := asyncpostgres.BindExternalPreEffect(store)
	if err != nil {
		return nil, err
	}
	definitions := []struct {
		kind, operation string
		namespace       uuid.UUID
	}{
		{DashboardExportDeleteJobKind, "delete", dashboardExportDeleteEffectNamespace},
		{DashboardExportExpireJobKind, "expire", dashboardExportExpireEffectNamespace},
	}
	processors := make([]jobexecution.Processor, 0, len(definitions))
	for _, definition := range definitions {
		policy, resolveErr := jobs.Resolve(definition.kind, 1)
		if resolveErr != nil || !validDashboardExportJobPolicy(policy, definition.kind) {
			return nil, fmt.Errorf("Analytics dashboard export policy %s is invalid",
				definition.kind)
		}
		provider, buildErr := newDashboardExportJobProvider(
			definition.kind, definition.operation, worker, now)
		if buildErr != nil {
			return nil, buildErr
		}
		registry, buildErr := effectexecution.NewRegistry(
			[]effectexecution.Provider{provider})
		if buildErr != nil {
			return nil, buildErr
		}
		processor, buildErr := jobexecution.NewExternalProcessor(policy,
			dashboardExportJobPlanner{kind: definition.kind,
				operation: definition.operation, namespace: definition.namespace},
			dashboardExportPreEffect{store: preEffect}, store, registry)
		if buildErr != nil {
			return nil, buildErr
		}
		processors = append(processors, processor)
	}
	return processors, nil
}

type dashboardExportPreEffect struct {
	store asyncpostgres.ExternalPreEffectStore
}

func (value dashboardExportPreEffect) Start(ctx context.Context,
	claim asyncpostgres.JobClaim, spec asyncpostgres.ExternalEffectRequestSpec,
) (jobexecution.StartedEffect, error) {
	permit, err := value.store.Start(ctx, claim, spec)
	if err != nil {
		return jobexecution.StartedEffect{}, err
	}
	return jobexecution.NewStartedEffect(permit)
}

func (value dashboardExportPreEffect) FailBeforeEffect(ctx context.Context,
	claim asyncpostgres.JobClaim, failure asyncpostgres.JobFailure,
) (asyncpostgres.JobFailureResult, error) {
	return value.store.FailBeforeEffect(ctx, claim, failure)
}
