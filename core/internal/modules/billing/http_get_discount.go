package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) getDiscount(w http.ResponseWriter, r *http.Request) {
	access, err := requireAccess(handler, r, billingapi.PermissionPlatformBillingRead, operationDiscountRead)
	id, valid := pathUUID(r, "discount_id")
	if err != nil || !valid {
		handler.writeFailure(w, r, firstError(err, billingRequestInvalid))
		return
	}
	result, err := handler.catalog.GetDiscount(r.Context(), access, id)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, result)
}
