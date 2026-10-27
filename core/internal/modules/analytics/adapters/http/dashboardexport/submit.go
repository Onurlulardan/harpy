package dashboardexporthttp

import (
	"errors"
	"net/http"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

func (handler *Handler) Submit(writer http.ResponseWriter, request *http.Request) {
	access, err := handler.runtime.AccessFor(request,
		httptransport.PermissionDashboardExport, httptransport.OperationDashboardExportCreate)
	if err != nil {
		if errors.Is(err, exportapp.ErrQueueLimit) {
			writer.Header().Set("Retry-After", "5")
		}
		handler.runtime.Failure(writer, request, err)
		return
	}
	dashboardID, dashboardOK := httptransport.PathV7(request, "dashboard_id")
	key, keyOK := httptransport.IdempotencyKey(request)
	if !dashboardOK || !keyOK || request.URL.RawQuery != "" {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	document, err := httptransport.DecodeLimit[submitDocument](writer, request, 64<<10)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	result, err := handler.application.SubmitDashboardExport(request.Context(), access,
		exportapp.SubmitRequest{DashboardID: dashboardID, Format: exportapp.Format(document.Format),
			Parameters: document.Parameters, IdempotencyKey: key,
			Auth: exportapp.AuthEvidence{TokenVersion: access.TokenVersion(),
				AuthorizationVersion: access.AuthorizationVersion(),
				AuthorizationClock:   access.AuthorizationClock()}})
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	writer.Header().Set("Location", exportLocation(result))
	writer.Header().Set("Retry-After", "1")
	handler.runtime.JSON(writer, request, http.StatusAccepted, present(result))
}

func exportLocation(value exportapp.Snapshot) string {
	return "/api/analytics/dashboards/" + value.DashboardID.String() +
		"/exports/" + value.ExportID.String()
}
