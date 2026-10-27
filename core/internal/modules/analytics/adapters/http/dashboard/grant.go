package dashboardhttp

import (
	"net/http"

	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

func (handler *Handler) Grant(writer http.ResponseWriter, request *http.Request) {
	access, err := handler.runtime.AccessFor(request,
		httptransport.PermissionDashboardUpdate, httptransport.OperationDashboardGrantACL)
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
	document, err := httptransport.Decode[grantDocument](writer, request)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	principalID, principalOK := httptransport.WireUUID(document.PrincipalID)
	if !principalOK {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	result, err := handler.application.GrantDashboardACL(request.Context(), access, dashboardapp.GrantACLRequest{
		DashboardID: dashboardID, PrincipalID: principalID, PrincipalKind: document.PrincipalKind,
		Level: document.Level, ExpectedRevision: document.ExpectedRevision, IdempotencyKey: key})
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	handler.runtime.JSON(writer, request, http.StatusCreated, map[string]any{
		"dashboard_id": result.DashboardID.String(), "principal_id": result.PrincipalID.String(),
		"principal_kind": result.PrincipalKind, "level": result.Level, "revision": result.Revision})
}
