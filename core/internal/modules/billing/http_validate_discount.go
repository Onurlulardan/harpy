package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) validateDiscount(w http.ResponseWriter, r *http.Request) {
	access, err := requireAccess(handler, r, billingapi.PermissionBillingRead, operationDiscountValidate)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	input, decodeErr := decodeCommand[discountValidationInput](w, r)
	_, versionValid := parseUUID(input.OfferingVersionID)
	if decodeErr != nil || !requiredText(input.Code, 64) || !versionValid {
		handler.writeFailure(w, r, firstError(decodeErr, billingRequestInvalid))
		return
	}
	result, err := handler.tenant.ValidateDiscount(
		r.Context(), billingapi.ValidateDiscountCommand{
			Access: access, Code: input.Code, OfferingVersionID: input.OfferingVersionID,
		},
	)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, result)
}
