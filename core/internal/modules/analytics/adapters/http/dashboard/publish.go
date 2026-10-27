package dashboardhttp

import (
	"net/http"

	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

func (handler *Handler) Publish(writer http.ResponseWriter, request *http.Request) {
	access, err := handler.runtime.AccessFor(request,
		httptransport.PermissionDashboardPublish, httptransport.OperationDashboardPublish)
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
	document, err := httptransport.Decode[publishDocument](writer, request)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	draftID, draftOK := httptransport.WireV7(document.DraftVersionID)
	if !draftOK {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	result, err := handler.application.PublishDashboard(request.Context(), access, dashboardapp.PublishRequest{
		DashboardID: dashboardID, DraftVersionID: draftID, ExpectedRootRevision: document.ExpectedRootRevision,
		ExpectedDraftRevision: document.ExpectedDraftRevision, IdempotencyKey: key})
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	handler.runtime.JSON(writer, request, http.StatusCreated, present(result))
}
