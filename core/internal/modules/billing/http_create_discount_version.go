package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) createDiscountVersion(w http.ResponseWriter, r *http.Request) {
	access, err := requireAccess(handler, r, billingapi.PermissionPlatformBillingManage, operationDiscountPublish)
	id, idValid := pathUUID(r, "discount_id")
	key, keyValid := idempotencyKey(r)
	input, decodeErr := decodeCommand[discountVersionInput](w, r)
	if err != nil || !idValid || !keyValid || decodeErr != nil || !input.validMutation() {
		handler.writeFailure(w, r, firstError(err, firstError(decodeErr, billingRequestInvalid)))
		return
	}
	result, err := handler.catalog.CreateDiscountVersion(
		r.Context(), billingapi.CreateDiscountVersionCommand{
			Access: access, IdempotencyKey: key, DiscountID: id,
			ExpectedRevision: input.ExpectedRevision,
			Version:          input.application(),
		},
	)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusCreated, result)
}
