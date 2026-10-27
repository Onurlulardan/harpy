package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

type changePlanRequest struct {
	PlanID string `json:"plan_id"`
}

func (handler *Handler) changePlan(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.change-plan"
	access, err := handler.authorize(r, routeID, billingapi.PermissionBillingManage)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	input, err := decodeBody[changePlanRequest](r)
	if err != nil || !required(input.PlanID) {
		handler.fail(w, routeID, ErrRequestInvalid)
		return
	}
	current, err := handler.tenant.GetCurrentSubscription(
		r.Context(), billingapi.SubscriptionAccess{Access: access},
	)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	offerings, err := handler.publicOfferings(r.Context())
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	identity, err := handler.resolvePublicPlanIdentity(r.Context(), input.PlanID)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	target, err := findOffering(offerings, identity.CanonicalID)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	price, err := defaultPrice(target)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	key, err := handler.idempotencyKey(r)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	view, err := handler.tenant.ScheduleChange(r.Context(), billingapi.ScheduleChangeCommand{
		Access: access, IdempotencyKey: key, SubscriptionID: current.ID,
		TargetOfferingVersionID: target.VersionID, TargetPriceVersionID: price.ID,
		ExpectedVersion: current.Version,
	})
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	wire, err := handler.presentSubscription(r.Context(), view)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	writeJSON(w, http.StatusOK, wire)
}
