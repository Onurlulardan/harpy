package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) updatePlan(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.admin-plan-update"
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
	identity, err := handler.resolvePlanIdentity(r.Context(), access, id)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	current, err := handler.catalog.GetOffering(
		r.Context(), access, identity.CanonicalID,
	)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	if current.State != "active" {
		handler.fail(w, routeID, billingapi.ErrNotFound)
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
	command, err := mapPlanUpdate(access, key, current, input, handler.clock())
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	view, err := handler.catalog.CreateOfferingVersion(r.Context(), command)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	wire, err := planWithLegacyID(view, identity)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	writeJSON(w, http.StatusCreated, wire)
}
