package analytics

import (
	exportartifact "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/artifact/dashboardexport"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

type DashboardExportWorker = exportapp.Worker
type DashboardExportWorkerStore = exportapp.WorkerStore
type DashboardExportWorkerAuthorizer = exportapp.WorkerAuthorizer
type DashboardExportAuthorizationRequest = exportapp.AuthorizationRequest
type DashboardExportWidgetExecutor = exportapp.WidgetExecutor
type DashboardExportArtifactRuntime = exportartifact.Runtime
type DashboardExportContentRuntime = exportapp.ContentRuntime
type DashboardExportClock = exportapp.Clock

const (
	DashboardExportObjectPrefix           = "analytics-dashboard-exports"
	DashboardExportObjectContract         = "ANALYTICS-DASHBOARD-EXPORT-FCART1"
	DashboardExportMaximumCiphertextBytes = int64(128 << 20)
)

var (
	NewDashboardExportWorker          = exportapp.NewWorker
	NewDashboardExportRenderer        = exportapp.NewDeterministicRenderer
	NewDashboardExportArtifactRuntime = exportartifact.NewRuntime
	NewDashboardExportContentRuntime  = exportartifact.NewContentRuntime
)
