package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) resumeCancellation(w http.ResponseWriter, r *http.Request) {
	access, err := requireAccess(handler, r, billingapi.PermissionBillingManage, operationSubscriptionResume)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	key, valid := idempotencyKey(r)
	input, decodeErr := decodeCommand[resumeCancellationInput](w, r)
	if !valid || decodeErr != nil || input.ExpectedVersion < 1 {
		handler.writeFailure(w, r, firstError(decodeErr, billingRequestInvalid))
		return
	}
	result, err := handler.tenant.ResumeCancellation(
		r.Context(), billingapi.ResumeCancellationCommand{
			Access: access, IdempotencyKey: key, ExpectedVersion: input.ExpectedVersion,
		},
	)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, result)
}
