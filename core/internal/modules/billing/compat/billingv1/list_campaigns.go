package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) listCampaigns(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.admin-campaigns-list"
	access, err := handler.authorize(r, routeID, billingapi.PermissionPlatformBillingRead)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	readAccess, err := handler.catalogReadAccess(r, "billing.catalog.campaign.read")
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	views, err := handler.platformCampaigns(r.Context(), access, readAccess)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	wires := make([]campaignWire, len(views))
	for index, view := range views {
		wires[index], err = legacyCampaign(view)
		if err != nil {
			handler.fail(w, routeID, err)
			return
		}
	}
	writeJSON(w, http.StatusOK, wires)
}
