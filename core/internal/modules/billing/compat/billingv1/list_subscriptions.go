package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.admin-subscriptions-list"
	access, err := handler.authorize(r, routeID, billingapi.PermissionPlatformBillingRead)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	items, err := handler.allSubscriptions(r.Context(), access)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	result := make([]subscriptionWire, len(items))
	for index, view := range items {
		result[index], err = legacySubscription(view, view.Offering)
		if err != nil {
			handler.fail(w, routeID, err)
			return
		}
	}
	writeJSON(w, http.StatusOK, result)
}
