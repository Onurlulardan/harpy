package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) retireDiscount(w http.ResponseWriter, r *http.Request) {
	access, err := requireAccess(handler, r, billingapi.PermissionPlatformBillingManage, operationDiscountRetire)
	id, idValid := pathUUID(r, "discount_id")
	key, keyValid := idempotencyKey(r)
	input, decodeErr := decodeCommand[retireInput](w, r)
	if err != nil || !idValid || !keyValid || decodeErr != nil ||
		input.ExpectedRevision < 1 || !requiredText(input.Reason, 512) {
		handler.writeFailure(w, r, firstError(err, firstError(decodeErr, billingRequestInvalid)))
		return
	}
	result, err := handler.catalog.RetireDiscount(r.Context(), billingapi.RetireDiscountCommand{
		Access: access, IdempotencyKey: key, DiscountID: id,
		ExpectedRevision: input.ExpectedRevision, Reason: input.Reason,
	})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, result)
}
