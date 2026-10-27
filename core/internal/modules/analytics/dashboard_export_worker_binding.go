package analytics

import exportpostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/dashboardexport"

var BindDashboardExportWorkerStore = exportpostgres.BindWorker
