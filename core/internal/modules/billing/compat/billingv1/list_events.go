package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) listEvents(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.admin-events-list"
	access, err := handler.authorize(r, routeID, billingapi.PermissionPlatformBillingAudit)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	query, err := handler.eventReportQuery(access, r.URL.Query().Get("tenant_id"))
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	page, err := handler.reports.ListLegacyEvents(r.Context(), query)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	items := make([]eventWire, len(page.Items))
	for index, raw := range page.Items {
		view, materializeErr := materializeLegacyEvent(raw)
		if materializeErr != nil {
			handler.fail(w, routeID, materializeErr)
			return
		}
		items[index], err = legacyEvent(view)
		if err != nil {
			handler.fail(w, routeID, err)
			return
		}
	}
	writeJSON(w, http.StatusOK, items)
}
