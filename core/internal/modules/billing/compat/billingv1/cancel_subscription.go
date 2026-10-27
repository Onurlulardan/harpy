package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

type cancelRequest struct {
	Reason string `json:"reason"`
}

func (handler *Handler) cancelSubscription(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.cancel-subscription"
	access, err := handler.authorize(r, routeID, billingapi.PermissionBillingManage)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	input, err := decodeBody[cancelRequest](r)
	if err != nil {
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
	_, err = handler.tenant.ScheduleCancellation(
		r.Context(), billingapi.ScheduleCancellationCommand{
			Access: access, IdempotencyKey: key, SubscriptionID: current.ID,
			ExpectedVersion: current.Version, Reason: input.Reason,
		},
	)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Subscription will be cancelled at the end of the current billing period",
	})
}
