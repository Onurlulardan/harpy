package dashboardexporthttp

import (
	"net/http"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

func (handler *Handler) Delete(writer http.ResponseWriter, request *http.Request) {
	access, err := handler.runtime.AccessFor(request,
		httptransport.PermissionDashboardExport, httptransport.OperationDashboardExportDelete)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	dashboardID, dashboardOK := httptransport.PathV7(request, "dashboard_id")
	exportID, exportOK := httptransport.PathV7(request, "export_id")
	key, keyOK := httptransport.IdempotencyKey(request)
	if !dashboardOK || !exportOK || !keyOK || request.URL.RawQuery != "" ||
		!httptransport.BodyAbsent(request) {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	result, err := handler.application.DeleteDashboardExport(request.Context(), access,
		exportapp.DeleteRequest{DashboardID: dashboardID, ExportID: exportID,
			IdempotencyKey: key})
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	writer.Header().Set("Location", exportLocation(result))
	writer.Header().Set("Retry-After", "1")
	handler.runtime.JSON(writer, request, http.StatusAccepted, present(result))
}
