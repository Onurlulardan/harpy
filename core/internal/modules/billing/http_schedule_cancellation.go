package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) scheduleCancellation(w http.ResponseWriter, r *http.Request) {
	access, err := requireAccess(handler, r, billingapi.PermissionBillingManage, operationSubscriptionCancel)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	key, valid := idempotencyKey(r)
	input, decodeErr := decodeCommand[cancellationInput](w, r)
	if !valid || decodeErr != nil || input.ExpectedVersion < 1 || !requiredText(input.Reason, 512) {
		handler.writeFailure(w, r, firstError(decodeErr, billingRequestInvalid))
		return
	}
	result, err := handler.tenant.ScheduleCancellation(
		r.Context(), billingapi.ScheduleCancellationCommand{
			Access: access, IdempotencyKey: key,
			ExpectedVersion: input.ExpectedVersion, Reason: input.Reason,
		},
	)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusAccepted, result)
}
