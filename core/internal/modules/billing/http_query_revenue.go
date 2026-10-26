package billing

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) queryRevenue(w http.ResponseWriter, r *http.Request) {
	query, err := handler.platformQuery(
		w, r, billingapi.PermissionPlatformBillingRevenue, operationRevenueRead, true,
	)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	result, err := handler.reports.GetRevenue(r.Context(), query)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, result)
}
