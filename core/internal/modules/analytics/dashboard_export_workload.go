package analytics

import (
	"fmt"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

type DashboardExportAsyncWorkload = exportapp.AsyncWorkload
type DashboardExportWorkload = exportapp.Workload
type DashboardExportRecoveryRunner = exportapp.RecoveryRunner
type DashboardExportIDSource = exportapp.IDSource
type DashboardExportAdmission = exportapp.DispatcherAdmission
type DashboardExportSpoolAllocator = exportapp.SpoolAllocator

type dashboardExportRuntimeStore interface {
	exportapp.WorkerStore
	exportapp.RenderClaimStore
	exportapp.RecoveryStore
}

type dashboardExportRuntimeArtifacts interface {
	exportapp.ArtifactRuntime
	exportapp.SpoolAllocator
}

func NewDashboardExportWorkload(
	clock DashboardExportClock, ids DashboardExportIDSource,
	store dashboardExportRuntimeStore, authorizer DashboardExportWorkerAuthorizer,
	renderer exportapp.Renderer, artifacts dashboardExportRuntimeArtifacts,
	admission DashboardExportAdmission,
) (*DashboardExportWorkload, error) {
	workload, _, err := NewDashboardExportWorkerWorkload(
		clock, ids, store, authorizer, renderer, artifacts, admission,
	)
	return workload, err
}

func NewDashboardExportWorkerWorkload(
	clock DashboardExportClock, ids DashboardExportIDSource,
	store dashboardExportRuntimeStore, authorizer DashboardExportWorkerAuthorizer,
	renderer exportapp.Renderer, artifacts dashboardExportRuntimeArtifacts,
	admission DashboardExportAdmission,
) (*DashboardExportWorkload, *DashboardExportWorker, error) {
	worker, err := exportapp.NewWorker(
		store, authorizer, renderer, artifacts, clock, ids)
	if err != nil {
		return nil, nil, fmt.Errorf("build Analytics dashboard export worker: %w", err)
	}
	dispatcher, err := exportapp.NewRenderDispatcher(
		clock, ids, store, artifacts, worker, admission)
	if err != nil {
		return nil, nil, fmt.Errorf("build Analytics dashboard export dispatcher: %w", err)
	}
	recovery, err := exportapp.NewRecoveryRunner(
		clock, ids, store, artifacts, artifacts, admission)
	if err != nil {
		return nil, nil, fmt.Errorf("build Analytics dashboard export recovery: %w", err)
	}
	workload, err := exportapp.NewWorkload(dispatcher, recovery)
	if err != nil {
		return nil, nil, fmt.Errorf("build Analytics dashboard export workload: %w", err)
	}
	return workload, worker, nil
}
