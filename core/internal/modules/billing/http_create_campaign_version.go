package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) createCampaignVersion(w http.ResponseWriter, r *http.Request) {
	access, err := requireAccess(handler, r, billingapi.PermissionPlatformBillingManage, operationCampaignPublish)
	id, idValid := pathUUID(r, "campaign_id")
	key, keyValid := idempotencyKey(r)
	input, decodeErr := decodeCommand[campaignVersionInput](w, r)
	if err != nil || !idValid || !keyValid || decodeErr != nil || !input.validMutation() {
		handler.writeFailure(w, r, firstError(err, firstError(decodeErr, billingRequestInvalid)))
		return
	}
	result, err := handler.catalog.CreateCampaignVersion(
		r.Context(), billingapi.CreateCampaignVersionCommand{
			Access: access, IdempotencyKey: key, CampaignID: id,
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
