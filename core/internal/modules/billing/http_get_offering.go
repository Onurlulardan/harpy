package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) getOffering(w http.ResponseWriter, r *http.Request) {
	access, err := requireAccess(handler, r, billingapi.PermissionPlatformBillingRead, operationOfferingRead)
	id, valid := pathUUID(r, "offering_id")
	if err != nil || !valid {
		handler.writeFailure(w, r, firstError(err, billingRequestInvalid))
		return
	}
	result, err := handler.catalog.GetOffering(r.Context(), access, id)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, result)
}
