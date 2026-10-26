package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) scheduleChange(w http.ResponseWriter, r *http.Request) {
	access, err := requireAccess(handler, r, billingapi.PermissionBillingManage, operationSubscriptionChange)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	key, valid := idempotencyKey(r)
	input, decodeErr := decodeCommand[scheduleChangeInput](w, r)
	_, offeringValid := parseUUID(input.TargetOfferingVersionID)
	_, priceValid := parseUUID(input.TargetPriceVersionID)
	if !valid || decodeErr != nil || !offeringValid || !priceValid || input.ExpectedVersion < 1 {
		handler.writeFailure(w, r, firstError(decodeErr, billingRequestInvalid))
		return
	}
	result, err := handler.tenant.ScheduleChange(r.Context(), billingapi.ScheduleChangeCommand{
		Access: access, IdempotencyKey: key,
		TargetOfferingVersionID: input.TargetOfferingVersionID,
		TargetPriceVersionID:    input.TargetPriceVersionID,
		ExpectedVersion:         input.ExpectedVersion,
	})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusAccepted, result)
}
