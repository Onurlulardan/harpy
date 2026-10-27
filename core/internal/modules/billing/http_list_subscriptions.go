package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	query, err := handler.platformQuery(
		w, r, billingapi.PermissionPlatformBillingRead, operationSubscriptionList, false,
	)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	result, err := handler.reports.ListSubscriptions(r.Context(), query)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, result)
}
