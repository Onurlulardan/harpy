package dashboardhttp

import (
	"net/http"

	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

func (handler *Handler) Revoke(writer http.ResponseWriter, request *http.Request) {
	access, err := handler.runtime.AccessFor(request,
		httptransport.PermissionDashboardUpdate, httptransport.OperationDashboardRevokeACL)
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
	document, err := httptransport.Decode[revokeDocument](writer, request)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	principalID, principalOK := httptransport.WireUUID(document.PrincipalID)
	if !principalOK {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	err = handler.application.RevokeDashboardACL(request.Context(), access, dashboardapp.RevokeACLRequest{
		DashboardID: dashboardID, PrincipalID: principalID, PrincipalKind: document.PrincipalKind,
		ExpectedRevision: document.ExpectedRevision, IdempotencyKey: key})
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	handler.runtime.JSON(writer, request, http.StatusOK, map[string]bool{"revoked": true})
}
