package billingv1

import "net/http"

func (handler *Handler) getActiveCampaign(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.active-campaign"
	view, err := handler.public.GetActiveCampaign(r.Context())
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	if view == nil {
		writeJSON(w, http.StatusOK, nil)
		return
	}
	coupons := make([]couponWire, len(view.Discounts))
	for index, discount := range view.Discounts {
		coupons[index], err = legacyCoupon(discount)
		if err != nil {
			handler.fail(w, routeID, err)
			return
		}
	}
	campaign, err := legacyCampaign(*view)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"campaign": campaign, "coupons": coupons,
	})
}
