package dashboardhttp

import (
	"net/http"

	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	dashboarddomain "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/dashboard"
)

func (handler *Handler) Create(writer http.ResponseWriter, request *http.Request) {
	access, err := handler.runtime.AccessFor(request,
		httptransport.PermissionDashboardCreate, httptransport.OperationDashboardCreate)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	key, keyOK := httptransport.IdempotencyKey(request)
	if !keyOK || request.URL.RawQuery != "" {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	document, err := httptransport.Decode[createDocument](writer, request)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	items, itemsOK := widgets(document.Widgets)
	if !itemsOK {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	result, err := handler.application.CreateDashboard(request.Context(), access, dashboardapp.CreateRequest{
		Name: document.Name, Visibility: dashboarddomain.Visibility(document.Visibility),
		LayoutDocument: document.Layout, Widgets: items, IdempotencyKey: key})
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	writer.Header().Set("Location", dashboardsPath+"/"+result.DashboardID.String())
	handler.runtime.JSON(writer, request, http.StatusCreated, present(result))
}
