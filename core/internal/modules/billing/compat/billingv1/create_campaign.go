package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) createCampaign(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.admin-campaigns-create"
	access, err := handler.authorize(r, routeID, billingapi.PermissionPlatformBillingManage)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	input, err := decodeBody[campaignRequest](r)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	key, err := handler.idempotencyKey(r)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	command, err := mapCampaignCreate(access, key, input)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	view, err := handler.catalog.CreateCampaign(r.Context(), command)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	wire, err := legacyCampaign(view)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	writeJSON(w, http.StatusCreated, wire)
}
