package dashboardhttp

import (
	"net/http"

	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

func (handler *Handler) Set(writer http.ResponseWriter, request *http.Request) {
	access, err := handler.runtime.AccessFor(request,
		httptransport.PermissionDashboardRead, httptransport.OperationDashboardSetDefault)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	dashboardID, idOK := httptransport.PathV7(request, "dashboard_id")
	key, keyOK := httptransport.IdempotencyKey(request)
	surface := request.PathValue("surface")
	if !idOK || !keyOK || surface == "" || request.URL.RawQuery != "" || !httptransport.BodyAbsent(request) {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	err = handler.application.SetDefaultDashboard(request.Context(), access, dashboardapp.SetDefaultRequest{
		DashboardID: dashboardID, Surface: surface, IdempotencyKey: key})
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	handler.runtime.JSON(writer, request, http.StatusOK, map[string]bool{"default": true})
}
