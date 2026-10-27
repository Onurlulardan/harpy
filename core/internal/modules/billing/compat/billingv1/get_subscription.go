package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) getSubscription(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.subscription-read"
	access, err := handler.authorize(r, routeID, billingapi.PermissionBillingRead)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	view, err := handler.tenant.GetCurrentSubscription(
		r.Context(), billingapi.SubscriptionAccess{Access: access},
	)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	wire, err := handler.presentSubscription(r.Context(), view)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	writeJSON(w, http.StatusOK, wire)
}
