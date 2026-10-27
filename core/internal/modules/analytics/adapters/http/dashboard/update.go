package dashboardhttp

import (
	"net/http"

	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	dashboarddomain "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/dashboard"
)

func (handler *Handler) Update(writer http.ResponseWriter, request *http.Request) {
	access, err := handler.runtime.AccessFor(request,
		httptransport.PermissionDashboardUpdate, httptransport.OperationDashboardUpdate)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	dashboardID, idOK := httptransport.PathV7(request, "dashboard_id")
	key, keyOK := httptransport.IdempotencyKey(request)
	if !idOK || !keyOK || request.URL.RawQuery != "" {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	document, err := httptransport.Decode[updateDocument](writer, request)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	items, itemsOK := widgets(document.Widgets)
	if !itemsOK {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	result, err := handler.application.UpdateDashboard(request.Context(), access, dashboardapp.UpdateRequest{
		DashboardID: dashboardID, ExpectedRootRevision: document.ExpectedRootRevision,
		ExpectedVersionRevision: document.ExpectedVersionRevision, Name: document.Name,
		Visibility: dashboarddomain.Visibility(document.Visibility), LayoutDocument: document.Layout,
		Widgets: items, IdempotencyKey: key})
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	handler.runtime.JSON(writer, request, http.StatusOK, present(result))
}
