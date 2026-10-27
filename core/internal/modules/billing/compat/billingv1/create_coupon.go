package billingv1

import (
	"context"
	"errors"
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) createCoupon(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.admin-coupons-create"
	access, err := handler.authorize(r, routeID, billingapi.PermissionPlatformBillingManage)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	input, err := decodeBody[couponRequest](r)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	key, err := handler.idempotencyKey(r)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	references, err := handler.couponReferences(access)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	autoCode := input.Code == nil || *input.Code == ""
	view, err := handler.createCouponView(
		r.Context(), access, references, key, input, autoCode,
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

func (handler *Handler) createCouponView(
	ctx context.Context,
	access billingapi.Access,
	references couponReferenceAccess,
	key string,
	input couponRequest,
	autoCode bool,
) (billingapi.DiscountView, error) {
	for attempt := 0; attempt < 4; attempt++ {
		command, mapErr := handler.mapCouponCreate(
			ctx, access, references, key, input,
		)
		if mapErr != nil {
			return billingapi.DiscountView{}, mapErr
		}
		view, createErr := handler.catalog.CreateDiscount(ctx, command)
		if createErr != nil {
			if autoCode && errors.Is(createErr, billingapi.ErrConflict) && attempt < 3 {
				continue
			}
			return billingapi.DiscountView{}, createErr
		}
		return view, nil
	}
	return billingapi.DiscountView{}, billingapi.ErrConflict
}
