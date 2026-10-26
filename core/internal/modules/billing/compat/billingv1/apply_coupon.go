package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

type applyCouponRequest struct {
	CouponCode string `json:"coupon_code"`
}

func (handler *Handler) applyCoupon(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.apply-coupon"
	access, err := handler.authorize(r, routeID, billingapi.PermissionBillingManage)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	input, err := decodeBody[applyCouponRequest](r)
	if err != nil || !required(input.CouponCode) {
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
	key, err := handler.idempotencyKey(r)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	view, err := handler.tenant.RedeemDiscount(r.Context(), billingapi.RedeemDiscountCommand{
		Access: access, IdempotencyKey: key,
		SubscriptionID: current.ID, Code: input.CouponCode,
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
