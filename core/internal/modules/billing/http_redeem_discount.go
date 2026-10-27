package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) redeemDiscount(w http.ResponseWriter, r *http.Request) {
	access, err := requireAccess(handler, r, billingapi.PermissionBillingManage, operationDiscountRedeem)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	key, valid := idempotencyKey(r)
	input, decodeErr := decodeCommand[discountRedemptionInput](w, r)
	if !valid || decodeErr != nil || !requiredText(input.Code, 64) {
		handler.writeFailure(w, r, firstError(decodeErr, billingRequestInvalid))
		return
	}
	result, err := handler.tenant.RedeemDiscount(
		r.Context(), billingapi.RedeemDiscountCommand{
			Access: access, IdempotencyKey: key, Code: input.Code,
		},
	)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, result)
}
