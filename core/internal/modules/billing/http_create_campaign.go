package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) createCampaign(w http.ResponseWriter, r *http.Request) {
	access, err := requireAccess(handler, r, billingapi.PermissionPlatformBillingManage, operationCampaignCreate)
	key, keyValid := idempotencyKey(r)
	input, decodeErr := decodeCommand[campaignInput](w, r)
	if err != nil || !keyValid || decodeErr != nil || !requiredText(input.Key, 127) ||
		!input.InitialVersion.validInitial() {
		handler.writeFailure(w, r, firstError(err, firstError(decodeErr, billingRequestInvalid)))
		return
	}
	result, err := handler.catalog.CreateCampaign(r.Context(), billingapi.CreateCampaignCommand{
		Access: access, IdempotencyKey: key, Key: input.Key,
		InitialVersion: input.InitialVersion.application(),
	})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusCreated, result)
}
