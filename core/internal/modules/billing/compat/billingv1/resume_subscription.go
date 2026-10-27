package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) resumeSubscription(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.resume-subscription"
	access, err := handler.authorize(r, routeID, billingapi.PermissionBillingManage)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	if err := decodeOptionalBody[struct{}](r); err != nil {
		handler.fail(w, routeID, err)
		return
	}
	current, err := handler.tenant.GetCurrentSubscription(
		r.Context(), billingapi.SubscriptionAccess{Access: access},
	)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	key, err := handler.idempotencyKey(r)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	_, err = handler.tenant.ResumeCancellation(
		r.Context(), billingapi.ResumeCancellationCommand{
			Access: access, IdempotencyKey: key, SubscriptionID: current.ID,
			ExpectedVersion: current.Version,
		},
	)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Subscription resumed"})
}
