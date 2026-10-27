package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) createDiscount(w http.ResponseWriter, r *http.Request) {
	access, err := requireAccess(handler, r, billingapi.PermissionPlatformBillingManage, operationDiscountCreate)
	key, keyValid := idempotencyKey(r)
	input, decodeErr := decodeCommand[discountInput](w, r)
	if err != nil || !keyValid || decodeErr != nil || !input.InitialVersion.validInitial() {
		handler.writeFailure(w, r, firstError(err, firstError(decodeErr, billingRequestInvalid)))
		return
	}
	result, err := handler.catalog.CreateDiscount(r.Context(), billingapi.CreateDiscountCommand{
		Access: access, IdempotencyKey: key,
		InitialVersion: input.InitialVersion.application(),
	})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusCreated, result)
}
