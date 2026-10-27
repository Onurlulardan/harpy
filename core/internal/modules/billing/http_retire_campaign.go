package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) retireCampaign(w http.ResponseWriter, r *http.Request) {
	access, err := requireAccess(handler, r, billingapi.PermissionPlatformBillingManage, operationCampaignRetire)
	id, idValid := pathUUID(r, "campaign_id")
	key, keyValid := idempotencyKey(r)
	input, decodeErr := decodeCommand[retireInput](w, r)
	if err != nil || !idValid || !keyValid || decodeErr != nil ||
		input.ExpectedRevision < 1 || !requiredText(input.Reason, 512) {
		handler.writeFailure(w, r, firstError(err, firstError(decodeErr, billingRequestInvalid)))
		return
	}
	result, err := handler.catalog.RetireCampaign(r.Context(), billingapi.RetireCampaignCommand{
		Access: access, IdempotencyKey: key, CampaignID: id,
		ExpectedRevision: input.ExpectedRevision, Reason: input.Reason,
	})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, result)
}
