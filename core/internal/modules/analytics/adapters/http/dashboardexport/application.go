package dashboardexporthttp

import (
	"context"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	"github.com/google/uuid"
)

type Application interface {
	SubmitDashboardExport(context.Context, httptransport.Access,
		exportapp.SubmitRequest) (exportapp.Snapshot, error)
	GetDashboardExport(context.Context, httptransport.Access,
		uuid.UUID, uuid.UUID) (exportapp.Snapshot, error)
	DeleteDashboardExport(context.Context, httptransport.Access,
		exportapp.DeleteRequest) (exportapp.Snapshot, error)
	OpenDashboardExportContent(context.Context, httptransport.Access,
		uuid.UUID, uuid.UUID) (exportapp.Content, error)
}
