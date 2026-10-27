package dashboardhttp

import (
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

func (handler *Handler) List(writer http.ResponseWriter, request *http.Request) {
	access, err := handler.runtime.AccessFor(request,
		httptransport.PermissionDashboardRead, httptransport.OperationDashboardList)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	limit, after, inputOK := httptransport.PageInput(request)
	if !inputOK || !httptransport.BodyAbsent(request) {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	result, err := handler.application.ListDashboards(request.Context(), access, limit, after)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	handler.runtime.JSON(writer, request, http.StatusOK, presentPage(result))
}
