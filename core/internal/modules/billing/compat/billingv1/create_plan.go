package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) createPlan(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.admin-plans-create"
	access, err := handler.authorize(r, routeID, billingapi.PermissionPlatformBillingManage)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	input, err := decodeBody[planRequest](r)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	key, err := handler.idempotencyKey(r)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	command, err := mapPlanCreate(access, key, input, handler.clock())
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	view, err := handler.catalog.CreateOffering(r.Context(), command)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	writeJSON(w, http.StatusCreated, legacyPlan(view))
}
