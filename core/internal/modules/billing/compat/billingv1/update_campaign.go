package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) updateCampaign(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.admin-campaign-update"
	access, err := handler.authorize(r, routeID, billingapi.PermissionPlatformBillingManage)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	id, err := pathID(r)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	canonicalID, err := handler.canonicalPromotionID(
		r, access, billingapi.LegacyIdentityCampaign, id,
	)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	current, err := handler.catalog.GetCampaign(r.Context(), access, canonicalID)
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
	command, err := mapCampaignUpdate(access, key, current, input, handler.clock())
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	view, err := handler.catalog.CreateCampaignVersion(r.Context(), command)
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
