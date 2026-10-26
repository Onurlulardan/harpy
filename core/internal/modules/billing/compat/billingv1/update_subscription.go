package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

type subscriptionUpdateRequest struct {
	Status string  `json:"status"`
	PlanID *string `json:"plan_id"`
}

func (handler *Handler) updateSubscription(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.admin-subscription-update"
	access, err := handler.authorize(r, routeID, billingapi.PermissionPlatformBillingManage)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	input, err := decodeBody[subscriptionUpdateRequest](r)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	if input.PlanID != nil {
		handler.fail(w, routeID, ErrUnsupported)
		return
	}
	id, err := pathID(r)
	if err != nil || !validTransition(input.Status) {
		handler.fail(w, routeID, billingapi.ErrTransitionInvalid)
		return
	}
	current, err := handler.findSubscription(r, access, id)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	key, err := handler.idempotencyKey(r)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	transitioned, err := handler.reports.TransitionSubscription(
		r.Context(), billingapi.TransitionSubscriptionCommand{
			Access: access, IdempotencyKey: key, SubscriptionID: current.ID,
			ExpectedVersion: current.Version, TargetState: input.Status,
			Reason: "legacy API status transition",
		},
	)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	if transitioned.ID != current.ID {
		handler.fail(w, routeID, ErrSnapshotRequired)
		return
	}
	view, err := handler.findSubscription(r, access, id)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	wire, err := legacySubscription(view, view.Offering)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	writeJSON(w, http.StatusOK, wire)
}

func validTransition(value string) bool {
	switch value {
	case "trialing", "active", "past_due", "paused", "cancelled", "expired":
		return true
	default:
		return false
	}
}
