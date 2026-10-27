package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) listPlans(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.admin-plans-list"
	access, err := handler.authorize(r, routeID, billingapi.PermissionPlatformBillingRead)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	offerings, err := handler.platformOfferings(r.Context(), access)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	readAccess, err := handler.catalogReadAccess(r, "billing.catalog.offering.read")
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	plans, err := handler.presentAdminPlans(r.Context(), readAccess, offerings)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	writeJSON(w, http.StatusOK, plans)
}
