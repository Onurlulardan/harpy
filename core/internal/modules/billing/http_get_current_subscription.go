package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) getCurrentSubscription(w http.ResponseWriter, r *http.Request) {
	access, err := requireAccess(handler, r, billingapi.PermissionBillingRead, operationSubscriptionRead)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	result, err := handler.tenant.GetCurrentSubscription(
		r.Context(), billingapi.SubscriptionAccess{Access: access},
	)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, result)
}
