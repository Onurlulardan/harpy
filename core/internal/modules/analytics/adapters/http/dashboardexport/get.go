package dashboardexporthttp

import (
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

func (handler *Handler) Get(writer http.ResponseWriter, request *http.Request) {
	access, err := handler.runtime.AccessFor(request,
		httptransport.PermissionDashboardExport, httptransport.OperationDashboardExportGet)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	dashboardID, dashboardOK := httptransport.PathV7(request, "dashboard_id")
	exportID, exportOK := httptransport.PathV7(request, "export_id")
	if !dashboardOK || !exportOK || request.URL.RawQuery != "" || !httptransport.BodyAbsent(request) {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	result, err := handler.application.GetDashboardExport(
		request.Context(), access, dashboardID, exportID)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	handler.runtime.JSON(writer, request, http.StatusOK, present(result))
}
