package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) createOffering(w http.ResponseWriter, r *http.Request) {
	access, err := requireAccess(handler, r, billingapi.PermissionPlatformBillingManage, operationOfferingCreate)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	key, keyValid := idempotencyKey(r)
	input, decodeErr := decodeCommand[offeringInput](w, r)
	if !keyValid || decodeErr != nil || !requiredText(input.Key, 127) ||
		!input.InitialVersion.validInitial() {
		handler.writeFailure(w, r, firstError(decodeErr, billingRequestInvalid))
		return
	}
	result, err := handler.catalog.CreateOffering(r.Context(), billingapi.CreateOfferingCommand{
		Access: access, IdempotencyKey: key, Key: input.Key,
		InitialVersion: input.InitialVersion.application(),
	})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusCreated, result)
}
