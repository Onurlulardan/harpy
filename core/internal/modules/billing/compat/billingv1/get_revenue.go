package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) getRevenue(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.admin-revenue-read"
	access, err := handler.authorize(r, routeID, billingapi.PermissionPlatformBillingRevenue)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	query, err := handler.revenueReportQuery(access)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	result, err := handler.reports.GetRevenue(r.Context(), query)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
