package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) getPlan(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.admin-plan-read"
	access, err := handler.authorize(r, routeID, billingapi.PermissionPlatformBillingRead)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	id, err := pathID(r)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	identity, err := handler.resolvePlanIdentity(r.Context(), access, id)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	view, err := handler.catalog.GetOffering(r.Context(), access, identity.CanonicalID)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	if view.State != "active" {
		handler.fail(w, routeID, billingapi.ErrNotFound)
		return
	}
	wire, err := planWithLegacyID(view, identity)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	writeJSON(w, http.StatusOK, wire)
}
