package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) updateCoupon(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.admin-coupon-update"
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
		r, access, billingapi.LegacyIdentityDiscount, id,
	)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	current, err := handler.catalog.GetDiscount(r.Context(), access, canonicalID)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	input, err := decodeBody[couponRequest](r)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	references, err := handler.couponReferences(access)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	version, err := handler.couponVersion(
		r.Context(), current, input, current.Code, references,
	)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	key, err := handler.idempotencyKey(r)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	view, err := handler.catalog.CreateDiscountVersion(
		r.Context(), billingapi.CreateDiscountVersionCommand{
			Access: access, IdempotencyKey: key, DiscountID: current.ID,
			ExpectedRevision: current.Revision, Version: version,
		},
	)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	wire, err := legacyCoupon(view)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	writeJSON(w, http.StatusCreated, wire)
}
