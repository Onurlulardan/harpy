package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) createOfferingVersion(w http.ResponseWriter, r *http.Request) {
	access, err := requireAccess(handler, r, billingapi.PermissionPlatformBillingManage, operationOfferingPublish)
	id, idValid := pathUUID(r, "offering_id")
	key, keyValid := idempotencyKey(r)
	input, decodeErr := decodeCommand[offeringVersionInput](w, r)
	if err != nil || !idValid || !keyValid || decodeErr != nil || !input.validMutation() {
		handler.writeFailure(w, r, firstError(err, firstError(decodeErr, billingRequestInvalid)))
		return
	}
	result, err := handler.catalog.CreateOfferingVersion(
		r.Context(), billingapi.CreateOfferingVersionCommand{
			Access: access, IdempotencyKey: key, OfferingID: id,
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
