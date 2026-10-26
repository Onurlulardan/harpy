package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) getUsage(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.usage-read"
	access, err := handler.authorize(r, routeID, billingapi.PermissionBillingUsageRead)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	view, err := handler.tenant.GetUsage(
		r.Context(), billingapi.SubscriptionAccess{Access: access},
	)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	writeJSON(w, http.StatusOK, legacyUsage(view))
}
