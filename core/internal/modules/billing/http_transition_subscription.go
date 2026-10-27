package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) transitionSubscription(w http.ResponseWriter, r *http.Request) {
	access, err := requireAccess(handler, r, billingapi.PermissionPlatformBillingManage, operationSubscriptionTransition)
	id, idValid := pathUUID(r, "subscription_id")
	key, keyValid := idempotencyKey(r)
	input, decodeErr := decodeCommand[transitionInput](w, r)
	if err != nil || !idValid || !keyValid || decodeErr != nil ||
		input.ExpectedVersion < 1 || !validSubscriptionState(input.TargetState) ||
		!requiredText(input.Reason, 512) {
		handler.writeFailure(w, r, firstError(err, firstError(decodeErr, billingRequestInvalid)))
		return
	}
	result, err := handler.reports.TransitionSubscription(
		r.Context(), billingapi.TransitionSubscriptionCommand{
			Access: access, IdempotencyKey: key, SubscriptionID: id,
			ExpectedVersion: input.ExpectedVersion,
			TargetState:     input.TargetState, Reason: input.Reason,
		},
	)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, result)
}

func validSubscriptionState(value string) bool {
	switch value {
	case "trialing", "active", "past_due", "paused", "cancelled", "expired":
		return true
	default:
		return false
	}
}
